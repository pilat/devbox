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
		RunE: runWrapperWithProject(func(ctx context.Context, p *project.Project, cmd *cobra.Command, args []string) error {
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

	root.AddCommand(cmd)
}

func runMount(ctx context.Context, p *project.Project, sourceName, targetPath string) ([]string, error) {
	if sourceName == "" {
		_, s, err := manager.Autodetect()
		if err != nil {
			return nil, fmt.Errorf("failed to autodetect source name: %w", err)
		} else {
			sourceName = s
		}
	}

	affectedServices, err := p.Mount(ctx, sourceName, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to mount source code: %w", err)
	}

	return affectedServices, nil
}
