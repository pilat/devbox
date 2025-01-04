package composer

import (
	"context"
	"fmt"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

func Restart(ctx context.Context, project *types.Project, services []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	composer, err := getClient()

	psOpts := api.PsOptions{
		Project:  project,
		Services: services,
		All:      true,
	}

	info, err := composer.Ps(ctx, project.Name, psOpts)
	if err != nil {
		return fmt.Errorf("failed to get services info: %w", err)
	}

	// If any of the services is not in the list we are probably not running
	notInList := false
	for _, service := range services {
		found := false
		for _, s := range info {
			name := getServiceName(s.Name, project)
			if name == service {
				found = true
				break
			}
		}

		if !found {
			notInList = true
			break
		}
	}

	if notInList {
		return nil
	}

	restartOpts := api.RestartOptions{
		Services: services,
	}

	if err = composer.Restart(ctx, project.Name, restartOpts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
