package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/pkg/utils"
	"github.com/pilat/devbox/internal/planner"
)

func (c *cli) Start(name string) error {
	c.log.Debug("Update project", "name", name)
	err := c.update(name)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	c.log.Debug("Connect to docker")
	d, err := docker.New()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer d.Close()

	c.log.Debug("Ping docker")
	err = d.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("failed to ping docker: %w", err)
	}

	c.log.Debug("Get home dir")
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	targetPath := filepath.Join(homeDir, appFolder, name)

	c.log.Debug("Reading configuration", "target", targetPath)
	cfg, err := config.New(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	cfg.Name = name

	// stateFile := filepath.Join(homeDir, appFolder, name, ".devboxstate")
	// state, err := state.New(stateFile)
	// if err != nil {
	// 	return fmt.Errorf("failed to read state: %w", err)
	// }

	// TODO: override mounts
	// for _, source := range cfg.Sources {
	// 	if localPath, ok := state.Mounts[source.Name]; ok {
	// 		source.LocalPath = localPath
	// 	}
	// }
	// _ = state

	c.log.Debug("Start planner")
	err = planner.Start(context.Background(), d, c.log, cfg)
	if err != nil {
		return fmt.Errorf("failed to start planner: %w", err)
	}

	return nil
}
