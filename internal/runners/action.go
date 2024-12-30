package runners

import (
	"context"
	"log/slog"
	"time"

	"fmt"
	"strings"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/pkg/utils"
)

type actionRunner struct {
	cli docker.Service
	log *slog.Logger

	cfg    *config.Config
	action *config.ActionConfig

	dependsOn []string
}

var _ Runner = (*actionRunner)(nil)

func NewActionRunner(cli docker.Service, log *slog.Logger, cfg *config.Config, action *config.ActionConfig, dependsOn []string) Runner {
	return &actionRunner{
		cli: cli,
		log: log,

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

func (s *actionRunner) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled")
	}

	err := s.start(ctx)
	if err != nil {
		s.log.Error("Failed to start action", "error", err)
		return err
	}

	return nil
}

func (s *actionRunner) Stop(ctx context.Context) error {
	list, err := s.cli.ContainersList(ctx, docker.ContainersListOptions{
		All:     true,
		Filters: filterLabels(s.cfg.Name, "action", s.action.Name, ""),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %v", err)
	}

	for _, container := range list {
		timeout := 0
		stopOptions := docker.ContainerStopOptions{
			Timeout: &timeout,
		}
		if err := s.cli.ContainerStop(ctx, container.ID, stopOptions); err == nil {
			s.log.Debug("Container stopped", "container", container.ID)
		}

		err = s.cli.ContainerRemove(ctx, container.ID)
		if err != nil {
			return fmt.Errorf("failed to remove container: %v", err)
		}
	}

	return nil
}

func (s *actionRunner) start(ctx context.Context) error {
	networkConfig := &docker.NetworkNetworkingConfig{
		EndpointsConfig: map[string]*docker.NetworkEndpointSettings{
			s.cfg.NetworkName: {
				NetworkID: s.cfg.NetworkName,
			},
		},
	}

	mounts, err := getMounts(s.action.Volumes)
	if err != nil {
		return fmt.Errorf("failed to get mounts: %v", err)
	}

	hostConfig := &docker.ContainerHostConfig{
		Mounts: mounts,
	}

	env, err := getEnvs(s.action.Environment, s.action.EnvFile)
	if err != nil {
		return fmt.Errorf("failed to get envs: %v", err)
	}

	for i, cmds := range s.action.Commands {
		lastCmd := strings.Join(cmds, " ")

		containerName := fmt.Sprintf("%s-%s-%d", s.cfg.Name, s.action.Name, i)

		list, err := s.cli.ContainersList(ctx, docker.ContainersListOptions{
			All:     true,
			Filters: filterLabels(s.cfg.Name, "action", s.action.Name, containerName),
		})
		if err != nil {
			return fmt.Errorf("failed to list containers: %v", err)
		}
		if len(list) > 0 {
			s.log.Warn("Container already exists", "container", containerName)
			continue
		}

		hostname, err := utils.ConvertToRFCHostname(fmt.Sprintf("%s-%d", s.action.Name, i))
		if err != nil {
			s.log.Warn("Failed to convert hostname to RFC", "error", err)
			hostname = ""
		}

		s.log.Debug("Running action", "index", i, "command", lastCmd)
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
			return fmt.Errorf("failed to create container: %v", err)
		}

		s.log.Debug("Container created", "container_id", containerID)

		err = s.cli.ContainerStart(ctx, containerID)
		if err != nil {
			return fmt.Errorf("failed to start container: %v", err)
		}

		exitCode := 0
		err = func() error {
			s.log.Debug("Waiting for container to exit...")

			backoff := 50 * time.Millisecond

			deadline := time.Now().Add(time.Minute * 5)
			for time.Now().Before(deadline) {
				containerJSON, err := s.cli.ContainerInspect(ctx, containerID)
				if err != nil {
					return fmt.Errorf("failed to inspect container: %v", err)
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

		s.log.Debug("All commands completed", "exitCode", exitCode, "command", lastCmd)

		if exitCode != 0 {
			return fmt.Errorf(`last command "%s" failed with exit code %d`, lastCmd, exitCode)
		}
	}

	return nil
}
