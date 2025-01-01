package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/pkg/utils"
)

func (c *cli) update(name string) error {
	if name == "" {
		return fmt.Errorf("project name is required")
	}

	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	targetPath := filepath.Join(homeDir, appFolder, name)

	c.log.Debug("Checking if target folder exists", "folder", targetPath)
	_, err = os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("project %s does not exist", name)
	}

	git := git.New(targetPath)

	c.log.Debug("Running git pull", "folder", targetPath)
	err = git.Pull() // TODO: consider using git.Sync() to reset it every time
	if err != nil {
		return fmt.Errorf("failed to pull git repo: %w", err)
	}

	c.log.Debug("Get latest commit info")
	info, err := git.GetInfo()

	c.log.Info("Project updated",
		"name", name,
		"commit", info.Hash,
		"author", info.Author,
		"date", info.Date,
		"message", info.Message,
	)

	return nil
}
