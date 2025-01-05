package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pilat/devbox/internal/composer"
)

func (a *app) Mount(sourceName, path string) error {
	if sourceName == "" {
		_, s, err := a.autodetect()
		if err != nil {
			return fmt.Errorf("failed to autodetect source name: %w", err)
		} else {
			sourceName = s
		}
	}

	if path == "" {
		curDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		path = curDir
	}

	if _, ok := a.state.Mounts[sourceName]; ok {
		return fmt.Errorf("source %s already mounted", sourceName)
	}

	a.state.Mounts[sourceName] = path

	if err := a.state.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	ctx := context.TODO()

	fullPathToSources := filepath.Join(a.projectPath, sourcesDir, sourceName)
	affectedServices := a.getAffectedServices(fullPathToSources)

	if err := a.LoadProject(); err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	if err := a.restartServices(ctx, affectedServices); err != nil {
		return fmt.Errorf("failed to restart services: %w", err)
	}

	return nil
}

func (a *app) restartServices(ctx context.Context, affectedServices []string) error {
	if len(affectedServices) == 0 {
		return fmt.Errorf("no services affected what is not expected")
	}

	err := composer.Restart(ctx, a.project, affectedServices)
	if err != nil {
		return fmt.Errorf("failed to restart services: %w", err)
	}

	return nil
}
