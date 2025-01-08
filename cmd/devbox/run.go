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
		Use:   "run <scenario>",
		Short: "Run scenario defined in devbox project",
		Long:  "You can pass additional arguments to the scenario",
		Args:  cobra.MinimumNArgs(1),
		RunE: runWrapperWithProject(func(ctx context.Context, p *project.Project, cmd *cobra.Command, args []string) error {
			command := args[0]
			if len(args) > 1 {
				args = args[1:]
			} else {
				args = []string{}
			}

			if err := runRun(ctx, p, command, args); err != nil {
				return fmt.Errorf("failed to run scenario: %w", err)
			}

			return nil
		}),
	}

	cmd.Flags().SetInterspersed(false)

	cmd.ValidArgsFunction = validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, p *project.Project, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// If a scenario is already provided (or project is not detected), disallow further completions
		if len(args) > 0 || p == nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return p.GetScenarios(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runRun(ctx context.Context, p *project.Project, command string, args []string) error {
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

	if err := cli.Run(ctx, p, command, args); err != nil {
		return fmt.Errorf("failed to run scenario: %w", err)
	}

	return nil
}
