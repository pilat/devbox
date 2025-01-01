package cli

import (
	"fmt"
	"path/filepath"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/pkg/utils"
)

func (c *cli) Unmount(name, sourceName string) error {
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

	if _, ok := cfg.State.Mounts[sourceName]; !ok {
		return fmt.Errorf("source %s is not mounted", sourceName)
	}

	delete(cfg.State.Mounts, sourceName)

	err = cfg.State.Save()
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}
