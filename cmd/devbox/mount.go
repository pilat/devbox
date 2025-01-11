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
				if detectedName, _ := manager.AutodetectSource(p); detectedName != "" {
					sourceName = detectedName
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

			if sourceName == "" {
				if detectedName, _ := manager.AutodetectSource(p); detectedName != "" {
					sourceName = detectedName
				}
			}

			affectedServices, err := runMount(ctx, p, sourceName, targetPath)
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
	cmd.PersistentFlags().StringVarP(&targetPath, "path", "p", "", "Path to mount")

	_ = cmd.RegisterFlagCompletionFunc("source", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		p, err := manager.AutodetectProject(projectName)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return p.GetLocalMountCandidates(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runMount(ctx context.Context, p *project.Project, sourceName, targetPath string) ([]string, error) {
	// if sourceName == "" {
	// 	_, s, err := manager.Autodetect()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to autodetect source name: %w", err)
	// 	} else {
	// 		sourceName = s
	// 	}
	// }

	affectedServices, err := p.Mount(ctx, sourceName, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to mount source code: %w", err)
	}

	return affectedServices, nil
}
