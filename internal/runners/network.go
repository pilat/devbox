package runners

import (
	"context"
	"log/slog"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
)

type networkRunner struct {
	cli docker.Service
	log *slog.Logger

	cfg       *config.Config
	dependsOn []string
}

var _ Runner = (*networkRunner)(nil)

func NewNetworkRunner(cli docker.Service, log *slog.Logger, cfg *config.Config, dependsOn []string) Runner {
	return &networkRunner{
		cli: cli,
		log: log,

		cfg:       cfg,
		dependsOn: dependsOn,
	}
}

func (s *networkRunner) Ref() string {
	return s.cfg.NetworkName
}

func (s *networkRunner) DependsOn() []string {
	return s.dependsOn
}

func (s *networkRunner) Start(ctx context.Context) error {
	err := s.start(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *networkRunner) Stop(ctx context.Context) error {
	items, err := s.cli.ListNetworks(ctx, docker.NetworksListOptions{
		Filters: filterLabels(s.cfg.Name, "service", s.cfg.NetworkName, s.cfg.NetworkName),
	})
	if err != nil {
		return err
	}

	for _, item := range items {
		err = s.cli.DeleteNetwork(ctx, item.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *networkRunner) start(ctx context.Context) error {
	items, err := s.cli.ListNetworks(ctx, docker.NetworksListOptions{
		Filters: filterLabels(s.cfg.Name, "network", s.cfg.NetworkName, s.cfg.NetworkName),
	})
	if err != nil {
		return err
	}

	if len(items) > 0 {
		s.log.Info("Network already exists", "name", s.cfg.NetworkName)
		return nil
	}

	return s.cli.CreateNetwork(ctx, s.cfg.NetworkName, docker.NetworkCreateOptions{
		Labels: makeLabels(s.cfg.Name, "network", s.cfg.NetworkName),
	})
}
