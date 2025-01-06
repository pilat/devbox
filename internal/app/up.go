package app

import (
	"context"
	"fmt"
)

func (a *app) Up() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	ctx := context.TODO()

	if err := a.project.Validate(); err != nil {
		return fmt.Errorf("failed to validate project: %w", err)
	}

	if err := a.project.Up(ctx); err != nil {
		return fmt.Errorf("failed to up services: %w", err)
	}

	return nil
}
