package app

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/composer"
)

func (a *app) Down() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	ctx := context.TODO()

	err := composer.Down(ctx, a.project)
	if err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
