package app

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/git"
)

func (c *app) update() error {
	if c.projectPath == "" {
		return ErrProjectIsNotSet
	}

	git := git.New(c.projectPath)

	err := git.Pull(context.TODO()) // TODO: consider using git.Sync() to reset it every time
	if err != nil {
		return fmt.Errorf("failed to pull git repo: %w", err)
	}

	_, err = git.GetInfo(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get git info: %w", err)
	}

	return nil
}
