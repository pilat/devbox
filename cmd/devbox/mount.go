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
	var targetPath string

	cmd := &cobra.Command{
		Use:   "mount",
		Short: "Mount source code",
		Long:  "That command will mount source code to the project",
		Args:  cobra.MinimumNArgs(0),
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			if sourceName == "" {
				if sources, _, _ := manager.AutodetectSource(p, "", manager.AutodetectSourceForMount); len(sources) > 0 {
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				}
			}

			if sourceName == "" && !cmd.Flags().Changed("source") {
				return []string{"--source"}, cobra.ShellCompDirectiveNoFileComp
			}

			// We are not suggesting --path since we assume that user wants to mount the current directory

			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			sources, affectedServices, err := manager.AutodetectSource(p, sourceName, manager.AutodetectSourceForMount)
			if err != nil {
				return fmt.Errorf("failed to autodetect source: %w", err)
			}

			if err := runMount(ctx, p, sources, targetPath); err != nil {
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
	cmd.PersistentFlags().StringVarP(&targetPath, "path", "p", "", "Path to mount")

	_ = cmd.RegisterFlagCompletionFunc("source", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		p, err := manager.AutodetectProject(projectName)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return manager.GetLocalMountCandidates(p, toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runMount(ctx context.Context, p *project.Project, sources []string, targetPath string) error {
	err := p.Mount(ctx, sources, targetPath)
	if err != nil {
		return fmt.Errorf("failed to mount source code: %w", err)
	}

	return nil
}
