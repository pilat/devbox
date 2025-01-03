package runners

import (
	"context"

	"fmt"
	"time"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/pkg/utils"
)

type serviceRunner struct {
	cli docker.Service

	cfg     *config.Config
	service *config.ServiceConfig

	dependsOn []string
}

var _ Runner = (*serviceRunner)(nil)

func NewServiceRunner(cli docker.Service, cfg *config.Config, service *config.ServiceConfig, dependsOn []string) Runner {
	return &serviceRunner{
		cli: cli,

		cfg:     cfg,
		service: service,

		dependsOn: dependsOn,
	}
}

func (s *serviceRunner) Ref() string {
	return s.service.Name
}

func (s *serviceRunner) DependsOn() []string {
	return s.dependsOn
}

func (s *serviceRunner) Type() ServiceType {
	return TypeService
}

func (s *serviceRunner) Start(ctx context.Context) error {
	if err := s.start(ctx); err != nil {
		return fmt.Errorf("service '%s' failed: %w", s.service.Name, err)
	}

	return nil
}

func (s *serviceRunner) Stop(ctx context.Context) error {
	list, err := s.cli.ContainersList(ctx, docker.ContainersListOptions{
		All:     true,
		Filters: filterLabels(s.cfg.Name, "service", s.service.Name, ""),
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

func (s *serviceRunner) Destroy(ctx context.Context) error {
	return s.Stop(ctx)
}

func (s *serviceRunner) start(ctx context.Context) error {
	networkConfig := &docker.NetworkNetworkingConfig{
		EndpointsConfig: map[string]*docker.NetworkEndpointSettings{
			s.cfg.NetworkName: {
				NetworkID: s.cfg.NetworkName,
				Aliases:   s.service.HostAliases,
			},
		},
	}

	mounts, err := getMounts(s.cfg, s.service.Volumes)
	if err != nil {
		return fmt.Errorf("failed to get mounts: %w", err)
	}

	hostConfig := &docker.ContainerHostConfig{
		Mounts: mounts,
	}

	env, err := getEnvs(s.cfg.Name, s.service.Environment, s.service.EnvFile)
	if err != nil {
		return fmt.Errorf("failed to get envs: %w", err)
	}

	containerName := fmt.Sprintf("%s-%s", s.cfg.Name, s.service.Name)

	list, err := s.cli.ContainersList(ctx, docker.ContainersListOptions{
		All:     true,
		Filters: filterLabels(s.cfg.Name, "service", s.service.Name, containerName),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}
	if len(list) > 0 {
		return nil
	}

	hostname, err := utils.ConvertToRFCHostname(s.service.Name)
	if err != nil {
		hostname = ""
	}

	if s.service.Hostname != "" {
		hostname = s.service.Hostname
	}

	exposedPorts := make(map[docker.Port]struct{})
	bindings := docker.PortMap{}
	for _, port := range s.service.Ports {
		p := docker.Port(fmt.Sprintf("%d/%s", port.Target, port.Protocol))
		binding := docker.PortBinding{
			HostIP:   port.HostIP,
			HostPort: port.Published,
		}
		bindings[p] = append(bindings[p], binding)
		exposedPorts[p] = struct{}{}
	}
	hostConfig.PortBindings = bindings

	var healthcheck *docker.HealthConfig
	if s.service.Healthcheck != nil {
		healthcheck = &docker.HealthConfig{
			Test:     append([]string{"CMD"}, s.service.Healthcheck...), // docker and podman both support CMD
			Interval: 1 * time.Second,
			Timeout:  120 * time.Second,
		}
	}

	containerConfig := &docker.ContainerConfig{
		Image:        s.service.Image,
		Env:          env,
		WorkingDir:   s.service.WorkingDir,
		User:         s.service.User,
		Healthcheck:  healthcheck,
		Hostname:     hostname,
		Labels:       makeLabels(s.cfg.Name, "service", s.service.Name),
		ExposedPorts: exposedPorts,
	}

	if s.service.Entrypoint != nil {
		containerConfig.Entrypoint = *s.service.Entrypoint
	}

	if len(s.service.Command) > 0 {
		containerConfig.Cmd = s.service.Command
	}

	containerID, err := s.cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, containerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	err = s.cli.ContainerStart(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	err = func() error {
		deadline := time.Now().Add(time.Minute * 5)
		for time.Now().Before(deadline) {
			containerJSON, err := s.cli.ContainerInspect(ctx, containerID)
			if err != nil {
				return fmt.Errorf("failed to inspect container: %w", err)
			}

			health := containerJSON.State.Health
			if health == nil {
				return nil
			}

			if health.Status == "healthy" {
				return nil
			}

			time.Sleep(250 * time.Millisecond)
		}

		return fmt.Errorf("container did not become healthy within timeout: %s", containerID)
	}()

	if err != nil {
		return fmt.Errorf("failed to wait for container to become healthy: %w", err)
	}

	return nil
}
