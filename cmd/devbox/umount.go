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
		RunE: runWrapperWithProject(func(ctx context.Context, p *project.Project, cmd *cobra.Command, args []string) error {
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

	cmd.ValidArgsFunction = validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, p *project.Project, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if sourceName == "" {
			if _, s, err := manager.Autodetect(); err == nil {
				sourceName = s
			}
		}

		if sourceName == "" && !cmd.Flags().Changed("source") {
			return []string{"--source"}, cobra.ShellCompDirectiveNoFileComp
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.PersistentFlags().StringVarP(&sourceName, "source", "s", "", "Source name")

	_ = cmd.RegisterFlagCompletionFunc("source", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		p, err := getProject(context.Background())
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return p.GetLocalMounts(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runUmount(ctx context.Context, p *project.Project, sourceName string) ([]string, error) {
	if sourceName == "" {
		_, s, err := manager.Autodetect()
		if err != nil {
			return nil, fmt.Errorf("failed to autodetect source name: %w", err)
		} else {
			sourceName = s
		}
	}

	affectedServices, err := p.Umount(ctx, sourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to mount source code: %w", err)
	}

	return affectedServices, nil
}
