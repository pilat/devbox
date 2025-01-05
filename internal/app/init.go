package app

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/git"
)

func (a *app) Init(url string, branch string) error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	if a.isProjectExists() {
		return fmt.Errorf("project already exists")
	}

	fmt.Println(" Initializing project...")
	git := git.New(a.projectPath)
	err := git.Clone(context.TODO(), url, branch)
	if err != nil {
		return fmt.Errorf("failed to clone git repo: %w", err)
	}

	patterns := []string{
		"/sources/",
		"/.devboxstate",
	}

	err = git.SetLocalExclude(patterns)
	if err != nil {
		return fmt.Errorf("failed to set local exclude: %w", err)
	}

	if err := a.LoadProject(); err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	if err := a.UpdateSources(); err != nil {
		return fmt.Errorf("failed to update sources: %w", err)
	}

	return a.Info()
}
