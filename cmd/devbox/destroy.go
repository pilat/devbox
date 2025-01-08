package main

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy devbox project",
		Long:  "That command will destroy devbox project",
		RunE: runWrapperWithProject(func(ctx context.Context, p *project.Project, cmd *cobra.Command, args []string) error {
			if err := runDown(ctx, p, true); err != nil {
				return fmt.Errorf("failed to stop project: %w", err)
			}

			if err := runDestroy(ctx, p); err != nil {
				return fmt.Errorf("failed to destroy project: %w", err)
			}

			return nil
		}),
	}

	cmd.ValidArgsFunction = validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, p *project.Project, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runDestroy(ctx context.Context, p *project.Project) error {
	fmt.Println("[*] Removing project...")
	if err := manager.Destroy(ctx, p); err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}

	fmt.Println("")

	return nil
}
