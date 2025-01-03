package app

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/pkg/git"
)

func (a *app) UpdateProject() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	if !a.isProjectExists() {
		return fmt.Errorf("failed to get project path")
	}

	git := git.New(a.projectPath)

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
