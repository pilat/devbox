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
				if detectedName, _ := manager.AutodetectSource(p); detectedName != "" {
					sourceName = detectedName
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

			if sourceName == "" {
				if detectedName, _ := manager.AutodetectSource(p); detectedName != "" {
					sourceName = detectedName
				}
			}

			affectedServices, err := runUmount(ctx, p, sourceName)
			if err != nil {
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

		return p.GetLocalMounts(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runUmount(ctx context.Context, p *project.Project, sourceName string) ([]string, error) {
	affectedServices, err := p.Umount(ctx, sourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to mount source code: %w", err)
	}

	return affectedServices, nil
}
