package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func (a *app) Destroy() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	ctx := context.TODO()

	err := a.project.Down(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to down and remove volumes: %w", err)
	}

	projectPath := filepath.Join(a.homeDir, appFolder, a.project.Name)

	if !isProjectExists(projectPath) {
		return fmt.Errorf("project %s does not exist", a.project.Name)
	}

	fmt.Println(" Removing project...")
	err = os.RemoveAll(projectPath)
	if err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}

	fmt.Println("")

	return nil
}
