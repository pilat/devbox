package app

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/composer"
)

func (a *app) Up() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	ctx := context.TODO()

	if err := composer.Up(ctx, a.project); err != nil {
		return fmt.Errorf("failed to up services: %w", err)
	}

	return nil
}
