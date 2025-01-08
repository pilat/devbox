package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pilat/devbox/internal/project"
)

func (a *Service) Up(ctx context.Context, p *project.Project) error {
	timeout := 60 * time.Minute
	opts := project.UpOptions{
		Create: project.CreateOptions{
			RemoveOrphans: true,
			QuietPull:     true,
			Timeout:       &timeout,
			Inherit:       false,
		},
		Start: project.StartOptions{
			Project:     p.Project,
			Wait:        true,
			WaitTimeout: timeout,
		},
	}

	if err := a.service.Up(ctx, p.Project, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
