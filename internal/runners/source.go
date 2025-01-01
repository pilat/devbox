package runners

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/pkg/utils"
)

type sourceRunner struct {
	cli docker.Service
	log *slog.Logger

	appName   string
	src       config.SourceConfig
	dependsOn []string
}

var _ Runner = (*sourceRunner)(nil)

func NewSourceRunner(cli docker.Service, log *slog.Logger, appName string, src config.SourceConfig, dependsOn []string) Runner {
	return &sourceRunner{
		cli: cli,
		log: log,

		appName:   appName,
		src:       src,
		dependsOn: dependsOn,
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

func (s *sourceRunner) Destroy(ctx context.Context) error {
	return nil
}

func (s *sourceRunner) start(ctx context.Context) error {
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return err
	}

	targetPath := fmt.Sprintf("%s/.devbox/%s/sources/%s", homeDir, s.appName, s.src.Name)

	git := git.New(targetPath)
	err = git.Sync(s.src.URL, s.src.Branch, s.src.SparseCheckout)
	if err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	s.log.Debug("Get latest commit info")
	info, err := git.GetInfo()

	s.log.Info("Project updated",
		"name", s.src.Name,
		"commit", info.Hash,
		"author", info.Author,
		"date", info.Date,
		"message", info.Message,
	)

	return nil
}
