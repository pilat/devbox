package app

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/composer"
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

	err := composer.Build(ctx, a.project, services)
	if err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
