package runners

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
)

type sourceRunner struct {
	cli docker.Service
	log *slog.Logger

	appName              string
	provisionerImageName string
	src                  config.SourceConfig
	dependsOn            []string
}

var _ Runner = (*sourceRunner)(nil)

func NewSourceRunner(cli docker.Service, log *slog.Logger, appName, provisionerImageName string, src config.SourceConfig, dependsOn []string) Runner {
	return &sourceRunner{
		cli: cli,
		log: log,

		appName:              appName,
		provisionerImageName: provisionerImageName,
		src:                  src,
		dependsOn:            dependsOn,
	}
}

func (s *sourceRunner) Ref() string {
	return fmt.Sprintf("source.%s", s.src.Name)
}

func (s *sourceRunner) DependsOn() []string {
	return s.dependsOn
}

func (s *sourceRunner) Start(ctx context.Context) error {
	err := s.start(ctx)
	if err != nil {
		s.log.Error("Failed to start source", "error", err)
		return err
	}

	return nil
}

func (s *sourceRunner) Stop(ctx context.Context) error {
	return nil
}

func (s *sourceRunner) start(ctx context.Context) error {
	networkName := "default"

	// 1. Create volume
	err := s.cli.CreateVolume(ctx, docker.VolumeCreateOptions{
		Name: s.Ref(),
		Labels: map[string]string{
			"devbox":      "true",
			"devbox.type": "source",
			"devbox.name": s.appName,
		},
	})
	if err != nil {
		return err
	}

	// 2. Run temp container with volume mounted
	networkConfig := &docker.NetworkNetworkingConfig{
		EndpointsConfig: map[string]*docker.NetworkEndpointSettings{
			networkName: {
				NetworkID: networkName,
			},
		},
	}

	mounts := []docker.Mount{
		{
			Type:   docker.MountTypeVolume,
			Source: s.Ref(),
			Target: "/workspace",
			VolumeOptions: &docker.VolumeOptions{
				NoCopy: true,
			},
		},
	}

	hostConfig := &docker.ContainerHostConfig{
		Mounts:     mounts,
		AutoRemove: true,
	}

	env, err := getEnvs(s.src.Environment, s.src.EnvFile)
	if err != nil {
		return fmt.Errorf("failed to get envs: %v", err)
	}

	env = append(env, fmt.Sprintf("REPO_URL=%s", s.src.URL))
	env = append(env, fmt.Sprintf("TARGET_FOLDER=%s", s.src.Name))
	env = append(env, fmt.Sprintf("BRANCH_NAME=%s", s.src.Branch))
	env = append(env, fmt.Sprintf("SPARSE_CHECKOUT=%s", strings.Join(s.src.SparseCheckout, ",")))

	containerConfig := &docker.ContainerConfig{
		Cmd:        []string{"/bin/bash"},
		Entrypoint: []string{""},

		OpenStdin: true,

		Image: s.provisionerImageName,
		Env:   env,
	}

	containerID, err := s.cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	defer func() {
		// 4. Stop container
		timeout := 0
		stopOptions := docker.ContainerStopOptions{
			Timeout: &timeout,
		}
		_ = s.cli.ContainerStop(context.Background(), containerID, stopOptions)

		// 5. Remove container
		_ = s.cli.ContainerRemove(ctx, containerID)
	}()

	err = s.cli.ContainerStart(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	execID, err := s.cli.ContainerExecCreate(ctx, containerID, docker.ContainerExecOptions{
		Cmd:          []string{"/usr/local/bin/entrypoint.sh"},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	execResp, err := s.cli.ContainerExecAttach(ctx, execID, docker.ContainerExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}

	defer execResp.Close()

	var stdout, stderr bytes.Buffer
	done := make(chan error)

	go func() {
		_, err = docker.StdCopy(&stdout, &stderr, execResp.Reader)
		done <- err
	}()

	select {
	case <-done:
		break
	case <-ctx.Done():
		return fmt.Errorf("context cancelled")
	}

	// if strings.Contains(stderr.String(), "could not read Username") {
	// 	return ErrGithubTokenMissing
	// }

	lines := strings.Split(stdout.String(), "\n")
	if len(lines) < 4 {
		return fmt.Errorf("unexpected output: stdout=%q, stderr=%q", stdout.String(), stderr.String())
	}
	lines = lines[:4]

	s.log.Debug("Source downloaded",
		"repository", s.src.Name,
		"commit", lines[0],
		"author", lines[1],
		"date", lines[2],
		"message", lines[3],
	)

	return nil
}
