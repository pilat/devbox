package service

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/project"
)

func (a *Service) Down(ctx context.Context, p *project.Project, deleteVolumes bool) error {
	// we are not overriding timeout allowing users to define it with stop_grace_period by user
	opts := project.DownOptions{
		Project:       p.Project,
		RemoveOrphans: true,
		Volumes:       deleteVolumes,
	}

	if err := a.service.Down(ctx, p.Name, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
