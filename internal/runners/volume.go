package runners

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
)

type volumeRunner struct {
	cli docker.Service

	cfg       *config.Config
	volume    string
	dependsOn []string
}

var _ Runner = (*volumeRunner)(nil)

func NewVolumeRunner(cli docker.Service, cfg *config.Config, volume string, dependsOn []string) Runner {
	return &volumeRunner{
		cli: cli,

		cfg:       cfg,
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
	volumeName := fmt.Sprintf("%s-%s", s.cfg.Name, s.volume)

	err := s.cli.CreateVolume(ctx, docker.VolumeCreateOptions{
		Name: volumeName,
	})
	if err != nil {
		return fmt.Errorf("failed to create volume: %v", err)
	}

	return nil
}

func (s *volumeRunner) Stop(ctx context.Context) error {
	return nil
}

func (s *volumeRunner) Destroy(ctx context.Context) error {
	// TODO: remove volume

	return nil
}
