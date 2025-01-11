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
		Use:   "ps",
		Short: "List services in devbox project",
		Long:  "That command will list services in devbox project",
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if err := runPs(ctx, p); err != nil {
				return fmt.Errorf("failed to list services: %w", err)
			}

			return nil
		}),
	}

	root.AddCommand(cmd)
}

func runPs(ctx context.Context, p *project.Project) error {
	cli, err := service.New()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := cli.Ps(ctx, p); err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	return nil
}
