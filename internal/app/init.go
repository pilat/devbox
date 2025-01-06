package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pilat/devbox/internal/git"
)

func (a *app) Init(name, url string, branch string) error {
	if !validateName(name) {
		return fmt.Errorf("invalid project name: %s", name)
	}

	projectPath := filepath.Join(a.homeDir, appFolder, name)

	if isProjectExists(projectPath) {
		return fmt.Errorf("project already exists")
	}

	fmt.Println(" Initializing project...")
	git := git.New(projectPath)
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

	return nil
}
