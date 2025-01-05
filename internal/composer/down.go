package composer

import (
	"context"
	"fmt"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

func Down(ctx context.Context, project *types.Project) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	composer, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	// we are not overriding timeout allowing users to define it with stop_grace_period by user
	opts := api.DownOptions{
		Project:       project,
		RemoveOrphans: true,
	}

	fmt.Println("Down services...")
	if err = composer.Down(ctx, project.Name, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	fmt.Println("")

	return nil
}
