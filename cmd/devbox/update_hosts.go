package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	var cleanup bool

	cmd := &cobra.Command{
		Use:    "update-hosts",
		Hidden: true,
		Args:   cobra.MinimumNArgs(0),
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := mgr.AutodetectProject(ctx, projectName)
			if err != nil {
				return err
			}

			if err := runHostsUpdate(p, false, cleanup); err != nil {
				return fmt.Errorf("failed to update hosts file: %w", err)
			}

			return nil
		}),
	}

	cmd.Flags().BoolVarP(&cleanup, "cleanup", "c", false, "Cleanup")

	root.AddCommand(cmd)
}
