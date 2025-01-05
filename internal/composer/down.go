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

	opts := api.DownOptions{
		RemoveOrphans: true,
	}

	if err = composer.Down(ctx, project.Name, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
