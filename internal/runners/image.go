package runners

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
)

type imageRunner struct {
	cli docker.Service
	log *slog.Logger

	container *config.ContainerConfig // if nil, then it's an external image
	dependsOn []string
}

var _ Runner = (*imageRunner)(nil)

func NewImageRunner(cli docker.Service, log *slog.Logger, container *config.ContainerConfig, dependsOn []string) Runner {
	return &imageRunner{
		cli: cli,
		log: log,

		container: container,
		dependsOn: dependsOn,
	}
}

func (s *imageRunner) Ref() string {
	return s.container.Image
}

func (s *imageRunner) DependsOn() []string {
	return s.dependsOn
}

func (s *imageRunner) Start(ctx context.Context) error {
	err := s.start(ctx)
	if err != nil {
		s.log.Error("Failed to start image", "error", err)
		return err
	}

	return nil
}

func (s *imageRunner) Stop(ctx context.Context) error {
	return nil
}

func (s *imageRunner) start(ctx context.Context) error {
	s.log.Info("Creating Dockerfile", "image", s.container.Image)

	// Create tar archive containing Dockerfile and context directory
	contextBuffer := new(bytes.Buffer)
	tarWriter := tar.NewWriter(contextBuffer)
	defer tarWriter.Close()

	// Add Dockerfile to tar archive
	dockerfileHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(s.container.Dockerfile)),
		Mode: 0600,
	}
	err := tarWriter.WriteHeader(dockerfileHeader)
	if err != nil {
		return fmt.Errorf("failed to write Dockerfile header to tar: %v", err)
	}
	_, err = tarWriter.Write([]byte(s.container.Dockerfile))
	if err != nil {
		return fmt.Errorf("failed to write Dockerfile content to tar: %v", err)
	}

	// Set up context and build options
	buildOptions := docker.ImageBuildOptions{
		Tags:       []string{s.container.Image},
		Dockerfile: "Dockerfile",
		Version:    "2", // important for docker, not for podman
	}

	s.log.Debug("Starting Docker image build.", "image", s.container.Image)
	imageBuildResponse, err := s.cli.ImageBuild(ctx, contextBuffer, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build Docker image: %v", err)
	}
	defer imageBuildResponse.Body.Close()

	var stdout bytes.Buffer
	for {
		temp := make([]byte, 1024) // Temporary buffer to read one byte
		n, err := imageBuildResponse.Body.Read(temp)
		if err != nil {
			break
		}

		if n > 0 {
			_, _ = stdout.Write(temp[:n]) // Write the read byte to stdout
		}
	}

	s.log.Debug("Docker image build completed.", "image", s.container.Image, "output", stdout.String())
	return nil
}
