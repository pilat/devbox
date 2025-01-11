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
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if err := runDown(ctx, p, true); err != nil {
				return fmt.Errorf("failed to stop project: %w", err)
			}

			if err := runDestroy(ctx, p); err != nil {
				return fmt.Errorf("failed to destroy project: %w", err)
			}

			return nil
		}),
	}

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
