package app

import (
	"context"
	"fmt"
)

func (a *app) Down() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	ctx := context.TODO()

	err := a.project.Down(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
