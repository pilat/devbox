package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:    "install-ca",
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

			if err := runCertUpdate(p, false); err != nil {
				return fmt.Errorf("failed to install CA: %w", err)
			}

			return nil
		}),
	}

	root.AddCommand(cmd)
}
