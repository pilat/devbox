package main

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/project"
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
			p, err := mgr.AutodetectProject(ctx, projectName)
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
	// we are not overriding timeout allowing users to define it with stop_grace_period by user
	opts := project.DownOptions{
		Project:       p.Project,
		RemoveOrphans: true,
		Volumes:       deleteVolumes,
	}

	fmt.Println("[*] Down services...")
	if err := apiService.Down(ctx, p.Name, opts); err != nil {
		return fmt.Errorf("failed to stop project: %w", err)
	}
	fmt.Println("")

	return nil
}
