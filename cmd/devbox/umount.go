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
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			if sourceName == "" {
				if sources, _, _ := manager.AutodetectSource(p, ""); len(sources) > 0 {
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				}
			}

			if sourceName == "" && !cmd.Flags().Changed("source") {
				return []string{"--source"}, cobra.ShellCompDirectiveNoFileComp
			}

			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			sourceNameDetected, affectedServices, err := manager.AutodetectSource(p, sourceName)
			if err != nil {
				return fmt.Errorf("failed to autodetect source: %w", err)
			}

			if err := runUmount(ctx, p, sourceNameDetected); err != nil {
				return fmt.Errorf("failed to mount source code: %w", err)
			}

			if err := runRestart(ctx, p, affectedServices, false); err != nil {
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
		p, err := manager.AutodetectProject(projectName)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return manager.GetLocalMounts(p, toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runUmount(ctx context.Context, p *project.Project, detectedSource string) error {
	err := p.Umount(ctx, detectedSource)
	if err != nil {
		return fmt.Errorf("failed to mount source code: %w", err)
	}

	return nil
}
