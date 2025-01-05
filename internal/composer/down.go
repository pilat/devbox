package composer

import (
	"context"
	"fmt"
	"time"

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

	downTimeout := 1 * time.Second
	opts := api.DownOptions{
		Project:       project,
		RemoveOrphans: true,
		Timeout:       &downTimeout,
		// Services:      services,
	}

	fmt.Println("Down services...")
	if err = composer.Down(ctx, project.Name, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	fmt.Println("")

	return nil
}
