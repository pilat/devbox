package runners

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"

	"github.com/pilat/devbox/internal/docker"
)

type pullRunner struct {
	cli docker.Service
	log *slog.Logger

	name      string // only for external images
	dependsOn []string
}

var _ Runner = (*pullRunner)(nil)

func NewPullRunner(cli docker.Service, log *slog.Logger, image string) Runner {
	return &pullRunner{
		cli: cli,
		log: log,

		name:      image,
		dependsOn: []string{},
	}
}

func (s *pullRunner) Ref() string {
	return s.name
}

func (s *pullRunner) DependsOn() []string {
	return s.dependsOn
}

func (s *pullRunner) Start(ctx context.Context) error {
	err := s.start(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *pullRunner) Stop(ctx context.Context) error {
	return nil
}

func (s *pullRunner) start(ctx context.Context) error {
	// Check is image already exists
	_, err := s.cli.ImageInspectWithRaw(ctx, s.name)
	if err == nil {
		s.log.Debug("Image already exists", "image", s.name)
		return nil
	}

	// TODO: Introduce timeout because it may hang
	s.log.Debug("Pulling Docker image...")
	resp1, err := s.cli.ImagePull(ctx, s.name, docker.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
	}

	defer resp1.Close()

	scanner := bufio.NewScanner(resp1)
	for scanner.Scan() {
		line := scanner.Text()
		// s.log.Debug("Output", "line", line)
		_ = line
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read output: %v", err)
	}

	return nil
}
