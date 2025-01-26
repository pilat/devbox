package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/compose/v2/pkg/api"
	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	var profiles []string

	cmd := &cobra.Command{
		Use:               "restart",
		Short:             "Restart services in devbox project",
		Long:              "That command will restart services in devbox project",
		Args:              cobra.MinimumNArgs(0),
		ValidArgsFunction: validArgsWrapper(suggestRunningServices),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if err := runProjectUpdate(ctx, p); err != nil {
				return fmt.Errorf("failed to update project: %w", err)
			}

			if err := runHostsUpdate(p, true, false); err != nil {
				return fmt.Errorf("failed to update hosts file: %w", err)
			}

			if err := runCertUpdate(p, true); err != nil {
				return fmt.Errorf("failed to update certificates: %w", err)
			}

			if err := runSourcesUpdate(ctx, p); err != nil {
				return fmt.Errorf("failed to update sources: %w", err)
			}

			if err := p.Reload(ctx, profiles); err != nil {
				return fmt.Errorf("failed to reload project with profiles: %w", err)
			}

			if err := runRestart(ctx, p, args, true); err != nil {
				return fmt.Errorf("failed to restart services: %w", err)
			}

			return nil
		}),
	}

	cmd.PersistentFlags().StringSliceVarP(&profiles, "profile", "p", []string{}, "Profile to use")

	cmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		p, err := manager.AutodetectProject(projectName)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return getProfileCompletions(p, toComplete)
	})

	root.AddCommand(cmd)
}

func suggestRunningServices(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	p, err := manager.AutodetectProject(projectName)
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	// If project not detected, disallow further completions
	if p == nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	results, err := getRunningServices(ctx, apiService, p, false, toComplete)
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	return results, cobra.ShellCompDirectiveNoFileComp
}

func runRestart(ctx context.Context, p *project.Project, services []string, noDeps bool) error {
	isRunning, err := isRunning(ctx, apiService, p)
	if err != nil {
		return fmt.Errorf("failed to check if services are running: %w", err)
	}

	if noDeps == false && !isRunning {
		return nil
	}

	depOpt := project.IncludeDependents
	if noDeps { // in case of manual restart, we don't need to restart dependent services
		depOpt = project.IgnoreDependencies
	}

	projectWithServices, err := p.WithSelectedServices(services, depOpt)
	if err != nil {
		return fmt.Errorf("failed to select services: %w", err)
	}

	p = projectWithServices

	if isRunning {
		networksBackup := p.Networks
		p.Networks = project.Networks{} // to avoid an attempt to remove a network

		if err := runDown(ctx, p, false); err != nil {
			return err
		}

		p.Networks = networksBackup // network is needed for Up
	}

	if err := runBuild(ctx, p); err != nil {
		return err
	}

	if err := runUp(ctx, p); err != nil {
		return err
	}

	return nil
}

func getRunningServices(ctx context.Context, a api.Service, p *project.Project, all bool, filter string) ([]string, error) {
	opts := project.PsOptions{
		Project: p.Project,
		All:     all,
	}

	containers, err := a.Ps(ctx, p.Name, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	results := []string{}
	for _, container := range containers {
		containerName := container.Labels[project.ServiceLabel]
		if !strings.HasPrefix(strings.ToLower(containerName), strings.ToLower(filter)) {
			continue
		}

		results = append(results, containerName)
	}

	return results, nil
}
