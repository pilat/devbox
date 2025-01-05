package app

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/composer"
)

func (a *app) Restart(services []string, noDeps bool) error {
	ctx := context.TODO()

	isRunning, err := composer.IsRunning(ctx, a.project)
	if err != nil {
		return fmt.Errorf("failed to check if services are running: %w", err)
	}

	if !isRunning {
		return nil
	}

	depOpt := composer.IncludeDependents
	if noDeps { // in case of manual restart, we don't need to restart dependent services
		depOpt = composer.IgnoreDependencies
	}

	projectWithServices, err := a.project.WithSelectedServices(services, depOpt)
	if err != nil {
		return fmt.Errorf("failed to select services: %w", err)
	}

	a.project = projectWithServices

	networksBackup := a.project.Networks
	a.project.Networks = composer.Networks{} // to avoid an attempt to remove a network

	if err := a.Down(); err != nil {
		return err
	}

	a.project.Networks = networksBackup // network is needed for Up

	if err := a.Build(); err != nil {
		return err
	}

	if err := a.Up(); err != nil {
		return err
	}

	return nil
}
