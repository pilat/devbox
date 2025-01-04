package composer

import (
	"context"
	"fmt"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

func Up(ctx context.Context, project *types.Project) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	composer, err := getClient()

	timeout := 4 * time.Minute
	opts := api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
			QuietPull:     true,
			Timeout:       &timeout,
			Inherit:       false,
		},
		Start: api.StartOptions{
			Project:     project,
			Wait:        true,
			WaitTimeout: timeout,
		},
	}

	if err = composer.Up(ctx, project, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
