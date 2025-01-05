package composer

import (
	"context"
	"fmt"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

type serviceInfo struct {
	Name   string
	Age    string
	State  string
	Health string
}

func ListServices(ctx context.Context, project *types.Project) ([]serviceInfo, error) {
	composer, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	opts := api.PsOptions{
		Project: project,
		All:     true,
	}

	containers, err := composer.Ps(ctx, project.Name, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	results := make([]serviceInfo, 0, len(containers))
	for _, container := range containers {
		uptimeDuration := time.Since(time.Unix(container.Created, 0))

		h := uptimeDuration.Hours()
		days := int(h / 24)
		uptimeDuration -= time.Duration(days) * 24 * time.Hour
		hh := uptimeDuration.Hours()
		mm := uptimeDuration.Minutes()
		ss := uptimeDuration.Seconds()

		uptimeStr := ""
		if days > 0 {
			uptimeStr += fmt.Sprintf("%dd ", days)
		} else {
			uptimeStr += fmt.Sprintf("%02d:%02d:%02d", int(hh), int(mm)%60, int(ss)%60)
		}

		if container.State == "exited" {
			uptimeStr = ""
		}

		name := container.Labels[api.ServiceLabel]
		if name == "" {
			name = container.Name
		}

		results = append(results, serviceInfo{
			Name:   name,
			Age:    uptimeStr,
			State:  container.State,
			Health: container.Health,
		})
	}

	return results, nil
}

func IsRunning(ctx context.Context, project *types.Project) (bool, error) {
	composer, err := getClient()
	if err != nil {
		return false, fmt.Errorf("failed to get client: %w", err)
	}

	opts := api.PsOptions{
		Project: project,
	}

	containers, err := composer.Ps(ctx, project.Name, opts)
	if err != nil {
		return false, fmt.Errorf("failed to get services: %w", err)
	}

	hasAny := false
	for _, container := range containers {
		hasAny = container.Labels[api.ProjectLabel] == project.Name &&
			container.Labels[api.WorkingDirLabel] == project.WorkingDir
		if hasAny {
			break
		}
	}

	return hasAny, nil
}
