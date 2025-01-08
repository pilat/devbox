package service

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/project"
)

func (a *Service) Build(ctx context.Context, p *project.Project) error {
	uniqueImages := map[string]bool{}
	services := []string{}
	for _, service := range p.Services {
		if service.Build == nil || service.Image == "" {
			continue
		}

		if _, ok := uniqueImages[service.Image]; ok {
			continue
		}

		uniqueImages[service.Image] = true
		services = append(services, service.Name)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := project.BuildOptions{
		Services: services,
		Quiet:    true,
	}

	if err := a.service.Build(ctx, p.Project, opts); err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	return nil
}
