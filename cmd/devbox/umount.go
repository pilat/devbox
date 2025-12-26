package main

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	var sourceName string

	cmd := &cobra.Command{
		Use:   "umount",
		Short: "Umount source code",
		Long:  "That command will umount source code from the project",
		Args:  cobra.MinimumNArgs(0),
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			p, err := mgr.AutodetectProject(ctx, projectName)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			if sourceName == "" {
				if result, _ := mgr.AutodetectSource(ctx, p, "", manager.AutodetectSourceForUmount); result != nil && len(result.Sources) > 0 {
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				}
			}

			if sourceName == "" && !cmd.Flags().Changed("source") {
				return []string{"--source"}, cobra.ShellCompDirectiveNoFileComp
			}

			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := mgr.AutodetectProject(ctx, projectName)
			if err != nil {
				return err
			}

			result, err := mgr.AutodetectSource(ctx, p, sourceName, manager.AutodetectSourceForUmount)
			if err != nil {
				return fmt.Errorf("failed to autodetect source: %w", err)
			}

			if err := runUmount(ctx, p, result.Sources); err != nil {
				return fmt.Errorf("failed to unmount source code: %w", err)
			}

			if err := runRestart(ctx, p, result.AffectedServices, false); err != nil {
				return fmt.Errorf("failed to restart services: %w", err)
			}

			if err := runInfo(ctx, p); err != nil {
				return fmt.Errorf("failed to get project info: %w", err)
			}

			return nil
		}),
	}

	cmd.PersistentFlags().StringVarP(&sourceName, "source", "s", "", "Source name")

	_ = cmd.RegisterFlagCompletionFunc("source", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		p, err := mgr.AutodetectProject(context.Background(), projectName)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return mgr.GetLocalMounts(p, toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runUmount(ctx context.Context, p *project.Project, sources []string) error {
	err := p.Umount(ctx, sources)
	if err != nil {
		return fmt.Errorf("failed to unmount source code: %w", err)
	}

	return nil
}
