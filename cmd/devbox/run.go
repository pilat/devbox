package main

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/manager"
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
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			// If a scenario is already provided (or project is not detected), disallow further completions
			if len(args) > 0 || p == nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			return p.GetScenarios(toComplete), cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

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
