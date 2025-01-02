package app

import (
	"fmt"
	"os"
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

	a.state.Mounts[sourceName] = path

	if err := a.state.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// TODO:
	// What entities are directly affected to that mounts?
	// We should restart them and all services depending on them.

	return a.Info()
}
