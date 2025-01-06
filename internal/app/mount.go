package app

import (
	"fmt"
	"os"
	"path/filepath"
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

	fullPathToSources := filepath.Join(a.projectPath, sourcesDir, sourceName)
	affectedServices := a.servicesAffectedByMounts(fullPathToSources)

	if err := a.LoadProject(a.project.Name); err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	if err := a.Restart(affectedServices, false); err != nil {
		return fmt.Errorf("failed to restart services: %w", err)
	}

	return nil
}
