package composer

import (
	"context"
	"fmt"

	"github.com/docker/compose/v2/pkg/api"
)

func (p *Project) Down(ctx context.Context, deleteVolumes bool) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	composer, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	// we are not overriding timeout allowing users to define it with stop_grace_period by user
	opts := api.DownOptions{
		Project:       p.Project,
		RemoveOrphans: true,
		Volumes:       deleteVolumes,
	}

	fmt.Println("Down services...")
	if err = composer.Down(ctx, p.Name, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	fmt.Println("")

	return nil
}
