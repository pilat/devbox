package runners

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pilat/devbox/internal/docker"
)

type volumeRunner struct {
	cli docker.Service
	log *slog.Logger

	volume    string
	dependsOn []string
}

var _ Runner = (*volumeRunner)(nil)

func NewVolumeRunner(cli docker.Service, log *slog.Logger, volume string, dependsOn []string) Runner {
	return &volumeRunner{
		cli: cli,
		log: log,

		volume:    volume,
		dependsOn: dependsOn,
	}
}

func (s *volumeRunner) Ref() string {
	return s.volume
}

func (s *volumeRunner) DependsOn() []string {
	return s.dependsOn
}

func (s *volumeRunner) Start(ctx context.Context) error {
	err := s.start(ctx)
	if err != nil {
		s.log.Error("Failed to start volume", "error", err)
		return err
	}

	return nil
}

func (s *volumeRunner) Stop(ctx context.Context) error {
	return nil
}

func (s *volumeRunner) start(ctx context.Context) error {
	s.log.Info("Creating volume", "volume", s.volume)

	err := s.cli.CreateVolume(ctx, docker.VolumeCreateOptions{
		Name: s.volume,
	})
	if err != nil {
		return fmt.Errorf("failed to create volume: %v", err)
	}

	return nil
}
