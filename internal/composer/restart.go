package composer

import (
	"context"
	"fmt"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

func Restart(ctx context.Context, project *types.Project, services []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	composer, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	projectWithServices, err := project.WithSelectedServices(services)
	if err != nil {
		return fmt.Errorf("failed to select services: %w", err)
	}

	psOpts := api.PsOptions{
		Project: project,
		All:     true,
	}

	info, err := composer.Ps(ctx, project.Name, psOpts)
	if err != nil {
		return fmt.Errorf("failed to get services info: %w", err)
	}

	if len(info) == 0 {
		return nil // service is not running
	}

	downTimeout := 0 * time.Minute
	downOpts := api.DownOptions{
		Project:       projectWithServices,
		Timeout:       &downTimeout,
		Services:      services,
		RemoveOrphans: false,
	}

	if err := composer.Down(ctx, project.Name, downOpts); err != nil {
		return fmt.Errorf("failed to down selected services: %w", err)
	}

	upTimeout := 4 * time.Minute
	upOpts := api.UpOptions{
		Create: api.CreateOptions{
			QuietPull: true,
			Timeout:   &upTimeout,
			Inherit:   false,
		},
		Start: api.StartOptions{
			Project:     projectWithServices,
			Wait:        true,
			WaitTimeout: upTimeout,
		},
	}

	if err := composer.Up(ctx, projectWithServices, upOpts); err != nil {
		return fmt.Errorf("failed to up selected services: %w", err)
	}

	return nil
}
