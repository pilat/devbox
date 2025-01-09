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
		Use:   "logs",
		Short: "Show logs of services in devbox project",
		Long:  "That command will show logs of services in devbox project",
		Args:  cobra.MinimumNArgs(0),
		RunE: runWrapperWithProject(func(ctx context.Context, p *project.Project, cmd *cobra.Command, args []string) error {
			if err := runLogs(ctx, p, args); err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
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

		results, err := cli.GetRunningServices(ctx, p, true, toComplete)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return results, cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runLogs(ctx context.Context, p *project.Project, services []string) error {
	cli, err := service.New()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := cli.Logs(ctx, p, services); err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}
