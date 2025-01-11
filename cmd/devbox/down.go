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
		Use:   "down",
		Short: "Stop devbox project",
		Long:  "That command will stop devbox project",
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if err := runDown(ctx, p, false); err != nil {
				return fmt.Errorf("failed to stop project: %w", err)
			}

			return nil
		}),
	}

	root.AddCommand(cmd)
}

func runDown(ctx context.Context, p *project.Project, deleteVolumes bool) error {
	cli, err := service.New()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	fmt.Println("[*] Down services...")
	if err := cli.Down(ctx, p, deleteVolumes); err != nil {
		return fmt.Errorf("failed to stop project: %w", err)
	}
	fmt.Println("")

	return nil
}
