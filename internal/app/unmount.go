package app

import (
	"context"
	"fmt"
)

func (a *app) Unmount(sourceName string) error {
	if sourceName == "" {
		_, s, err := a.autodetect()
		if err != nil {
			return fmt.Errorf("failed to autodetect source name: %w", err)
		} else {
			sourceName = s
		}
	}

	curPath, ok := a.state.Mounts[sourceName]
	if !ok {
		return fmt.Errorf("source %s is not mounted", sourceName)
	}

	affectedServices := a.getAffectedServices(curPath)

	delete(a.state.Mounts, sourceName)

	if err := a.state.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	if err := a.LoadProject(); err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	ctx := context.TODO()
	if err := a.restartServices(ctx, affectedServices); err != nil {
		return fmt.Errorf("failed to restart services: %w", err)
	}

	return nil
}
