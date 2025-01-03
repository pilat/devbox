package runners

import (
	"bufio"
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/pkg/container"
)

type pullRunner struct {
	cli container.Service

	name      string // only for external images
	dependsOn []string
}

var _ Runner = (*pullRunner)(nil)

func NewPullRunner(cli container.Service, image string) Runner {
	return &pullRunner{
		cli: cli,

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

func (s *pullRunner) Type() ServiceType {
	return TypePull
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

func (s *pullRunner) Destroy(ctx context.Context) error {
	return nil
}

func (s *pullRunner) start(ctx context.Context) error {
	// Check is image already exists
	_, err := s.cli.ImageInspectWithRaw(ctx, s.name)
	if err == nil {
		return nil
	}

	// TODO: Introduce timeout because it may hang
	resp1, err := s.cli.ImagePull(ctx, s.name, container.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	defer resp1.Close()

	scanner := bufio.NewScanner(resp1)
	for scanner.Scan() {
		line := scanner.Text()
		_ = line
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read output: %w", err)
	}

	return nil
}
