package app

import (
	"fmt"
	"path/filepath"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/pkg/utils"
)

func (c *app) Mount(name, sourceName, path string) error {
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	targetPath := filepath.Join(homeDir, appFolder, name)

	cfg, err := config.New(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	cfg.Name = name

	// cfg.State.Mounts[sourceName] = path

	// err = cfg.State.Save()
	// if err != nil {
	// 	return fmt.Errorf("failed to save state: %w", err)
	// }

	return nil
}
