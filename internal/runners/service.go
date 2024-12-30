package runners

import (
	"context"
	"log/slog"

	"fmt"
	"time"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/pkg/utils"
)

type serviceRunner struct {
	cli docker.Service
	log *slog.Logger

	cfg     *config.Config
	service *config.ServiceConfig

	dependsOn []string
}

var _ Runner = (*serviceRunner)(nil)

func NewServiceRunner(cli docker.Service, log *slog.Logger, cfg *config.Config, service *config.ServiceConfig, dependsOn []string) Runner {
	return &serviceRunner{
		cli: cli,
		log: log,

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

func (s *serviceRunner) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled")
	}

	err := s.start(ctx)
	if err != nil {
		s.log.Error("Failed to start service", "name", s.service.Name, "error", err)
		return err
	}

	return nil
}

func (s *serviceRunner) Stop(ctx context.Context) error {
	list, err := s.cli.ContainersList(ctx, docker.ContainersListOptions{
		All:     true,
		Filters: filterLabels(s.cfg.Name, "service", s.service.Name, ""),
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

func (s *serviceRunner) start(ctx context.Context) error {
	s.log.Debug("Running service")

	networkConfig := &docker.NetworkNetworkingConfig{
		EndpointsConfig: map[string]*docker.NetworkEndpointSettings{
			s.cfg.NetworkName: {
				NetworkID: s.cfg.NetworkName,
				Aliases:   s.service.HostAliases,
			},
		},
	}

	mounts, err := getMounts(s.service.Volumes)
	if err != nil {
		return fmt.Errorf("failed to get mounts: %v", err)
	}

	hostConfig := &docker.ContainerHostConfig{
		Mounts: mounts,
	}

	env, err := getEnvs(s.service.Environment, s.service.EnvFile)
	if err != nil {
		return fmt.Errorf("failed to get envs: %v", err)
	}

	containerName := fmt.Sprintf("%s-%s", s.cfg.Name, s.service.Name)

	list, err := s.cli.ContainersList(ctx, docker.ContainersListOptions{
		All:     true,
		Filters: filterLabels(s.cfg.Name, "service", s.service.Name, containerName),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %v", err)
	}
	if len(list) > 0 {
		s.log.Debug("Container already exists", "container", containerName)
		return nil
	}

	hostname, err := utils.ConvertToRFCHostname(s.service.Name)
	if err != nil {
		s.log.Warn("Failed to convert hostname to RFC", "service", containerName, "error", err)
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
		return fmt.Errorf("failed to create container: %v", err)
	}

	s.log.Debug("Container created", "container_id", containerID)

	err = s.cli.ContainerStart(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	err = func() error {
		s.log.Debug("Waiting for container to become healthy...", "container", containerID)

		deadline := time.Now().Add(time.Minute * 5)
		for time.Now().Before(deadline) {
			containerJSON, err := s.cli.ContainerInspect(ctx, containerID)
			if err != nil {
				return fmt.Errorf("failed to inspect container: %v", err)
			}

			health := containerJSON.State.Health
			if health == nil {
				s.log.Warn("Health status not defined; we are considering it healthy", "container", containerID)
				return nil
			}

			if health.Status == "healthy" {
				s.log.Debug("Container is healthy", "container", containerID)
				return nil
			}

			time.Sleep(250 * time.Millisecond)
		}

		return fmt.Errorf("container did not become healthy within timeout: %s", containerID)
	}()

	if err != nil {
		return fmt.Errorf("failed to wait for container to become healthy: %v", err)
	}

	s.log.Debug("Service is running in background, leave it", "service", s.service.Name)

	return nil
}
