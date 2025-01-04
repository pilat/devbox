package composer

import (
	"context"
	"fmt"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

func Build(ctx context.Context, project *types.Project, services []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	composer, err := getClient()

	opts := api.BuildOptions{
		Services: services,
		Quiet:    true,
	}

	if err = composer.Build(ctx, project, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
