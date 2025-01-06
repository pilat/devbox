package composer

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/compose/v2/pkg/api"
)

func (p *Project) Up(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	composer, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	timeout := 4 * time.Minute
	opts := api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
			QuietPull:     true,
			Timeout:       &timeout,
			Inherit:       false,
		},
		Start: api.StartOptions{
			Project:     p.Project,
			Wait:        true,
			WaitTimeout: timeout,
		},
	}

	fmt.Println("Up services...")
	if err = composer.Up(ctx, p.Project, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	fmt.Println("")

	return nil
}
