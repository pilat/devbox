package main

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/project"
	"github.com/pilat/devbox/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart services in devbox project",
		Long:  "That command will restart services in devbox project",
		Args:  cobra.MinimumNArgs(0),
		RunE: runWrapperWithProject(func(ctx context.Context, p *project.Project, cmd *cobra.Command, args []string) error {
			if err := runProjectUpdate(ctx, p); err != nil {
				return fmt.Errorf("failed to update project: %w", err)
			}

			if err := runSourcesUpdate(ctx, p); err != nil {
				return fmt.Errorf("failed to update sources: %w", err)
			}

			if err := runRestart(ctx, p, args, true); err != nil {
				return fmt.Errorf("failed to restart services: %w", err)
			}

			return nil
		}),
	}

	cmd.ValidArgsFunction = validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, p *project.Project, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// If project not detected, disallow further completions
		if p == nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		cli, err := service.New()
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		results, err := cli.GetRunningServices(ctx, p, false, toComplete)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return results, cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runRestart(ctx context.Context, p *project.Project, services []string, noDeps bool) error {
	cli, err := service.New()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	isRunning, err := cli.IsRunning(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to check if services are running: %w", err)
	}

	if !isRunning {
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

	networksBackup := p.Networks
	p.Networks = project.Networks{} // to avoid an attempt to remove a network

	if err := runDown(ctx, p, false); err != nil {
		return err
	}

	p.Networks = networksBackup // network is needed for Up

	if err := runBuild(ctx, p); err != nil {
		return err
	}

	if err := runUp(ctx, p); err != nil {
		return err
	}

	return nil
}
