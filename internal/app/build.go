package app

import (
	"context"
	"fmt"
)

func (a *app) Build() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	ctx := context.TODO()

	uniqueImages := map[string]bool{}
	services := []string{}
	for _, service := range a.project.Services {
		if service.Build == nil || service.Image == "" {
			continue
		}

		if _, ok := uniqueImages[service.Image]; ok {
			continue
		}

		uniqueImages[service.Image] = true
		services = append(services, service.Name)
	}

	if err := a.project.Validate(); err != nil {
		return fmt.Errorf("failed to validate project: %w", err)
	}

	err := a.project.Build(ctx, services)
	if err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
