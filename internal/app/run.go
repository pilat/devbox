package app

import (
	"context"
	"fmt"
)

func (a *app) Run(command string, args []string) error {
	ctx := context.TODO()

	if err := a.project.Validate(); err != nil {
		return fmt.Errorf("failed to validate project: %w", err)
	}

	if err := a.project.Run(ctx, command, args); err != nil {
		return err
	}

	return nil
}
