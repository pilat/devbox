package composer

import (
	"context"
	"fmt"

	"github.com/docker/compose/v2/pkg/api"
)

func (p *Project) Build(ctx context.Context, services []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	composer, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	opts := api.BuildOptions{
		Services: services,
		Quiet:    true,
	}

	fmt.Println("Build services...")
	if err = composer.Build(ctx, p.Project, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	fmt.Println("")

	return nil
}
