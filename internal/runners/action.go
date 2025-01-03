package runners

import (
	"context"
	"time"

	"fmt"
	"strings"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/pkg/utils"
)

type actionRunner struct {
	cli docker.Service

	cfg    *config.Config
	action *config.ActionConfig

	dependsOn []string
}

var _ Runner = (*actionRunner)(nil)

func NewActionRunner(cli docker.Service, cfg *config.Config, action *config.ActionConfig, dependsOn []string) Runner {
	return &actionRunner{
		cli: cli,

		cfg:    cfg,
		action: action,

		dependsOn: dependsOn,
	}
}

func (s *actionRunner) Ref() string {
	return s.action.Name
}

func (s *actionRunner) DependsOn() []string {
	return s.dependsOn
}

func (s *actionRunner) Type() ServiceType {
	return TypeAction
}

func (s *actionRunner) Start(ctx context.Context) error {
	if err := s.start(ctx); err != nil {
		return fmt.Errorf("action '%s' has failed: %w", s.action.Name, err)
	}

	return nil
}

func (s *actionRunner) Stop(ctx context.Context) error {
	list, err := s.cli.ContainersList(ctx, docker.ContainersListOptions{
		All:     true,
		Filters: filterLabels(s.cfg.Name, "action", s.action.Name, ""),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range list {
		timeout := 0
		stopOptions := docker.ContainerStopOptions{
			Timeout: &timeout,
		}
		_ = s.cli.ContainerStop(ctx, container.ID, stopOptions)

		err = s.cli.ContainerRemove(ctx, container.ID)
		if err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}

	return nil
}

func (s *actionRunner) Destroy(ctx context.Context) error {
	return s.Stop(ctx)
}

func (s *actionRunner) start(ctx context.Context) error {
	networkConfig := &docker.NetworkNetworkingConfig{
		EndpointsConfig: map[string]*docker.NetworkEndpointSettings{
			s.cfg.NetworkName: {
				NetworkID: s.cfg.NetworkName,
			},
		},
	}

	mounts, err := getMounts(s.cfg, s.action.Volumes)
	if err != nil {
		return fmt.Errorf("failed to get mounts: %w", err)
	}

	hostConfig := &docker.ContainerHostConfig{
		Mounts: mounts,
	}

	env, err := getEnvs(s.cfg.Name, s.action.Environment, s.action.EnvFile)
	if err != nil {
		return fmt.Errorf("failed to get envs: %w", err)
	}

	for i, cmds := range s.action.Commands {
		lastCmd := strings.Join(cmds, " ")

		containerName := fmt.Sprintf("%s-%s-%d", s.cfg.Name, s.action.Name, i)

		list, err := s.cli.ContainersList(ctx, docker.ContainersListOptions{
			All:     true,
			Filters: filterLabels(s.cfg.Name, "action", s.action.Name, containerName),
		})
		if err != nil {
			return fmt.Errorf("failed to list containers: %w", err)
		}
		if len(list) > 0 {
			continue
		}

		hostname, err := utils.ConvertToRFCHostname(fmt.Sprintf("%s-%d", s.action.Name, i))
		if err != nil {
			hostname = ""
		}

		containerConfig := &docker.ContainerConfig{
			Image:      s.action.Image,
			Env:        env,
			WorkingDir: s.action.WorkingDir,
			User:       s.action.User,
			Hostname:   hostname,
			Labels:     makeLabels(s.cfg.Name, "action", s.action.Name),
		}

		if s.action.Entrypoint != nil {
			containerConfig.Entrypoint = *s.action.Entrypoint
		}

		if len(cmds) > 0 {
			containerConfig.Cmd = cmds
		}

		// Create container
		containerID, err := s.cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, containerName)
		if err != nil {
			return fmt.Errorf("failed to create container: %w", err)
		}

		err = s.cli.ContainerStart(ctx, containerID)
		if err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}

		exitCode := 0
		err = func() error {
			backoff := 50 * time.Millisecond

			deadline := time.Now().Add(time.Minute * 5)
			for time.Now().Before(deadline) {
				containerJSON, err := s.cli.ContainerInspect(ctx, containerID)
				if err != nil {
					return fmt.Errorf("failed to inspect container: %w", err)
				}

				exitCode = containerJSON.State.ExitCode

				if !containerJSON.State.Running {
					return nil
				}

				time.Sleep(backoff)

				backoff *= 2
				if backoff > 2*time.Second {
					backoff = 2 * time.Second
				}
			}

			return fmt.Errorf("container did not become healthy within timeout: %s", containerID)
		}()
		if err != nil {
			return err
		}

		if exitCode != 0 {
			return fmt.Errorf(`command "%s" failed with exit code %d`, lastCmd, exitCode)
		}
	}

	return nil
}
