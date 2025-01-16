package main

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/pilat/devbox/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:               "shell",
		Short:             "Run interactive shell in one of the services",
		Long:              "That command will run interactive shell in one of the services",
		Args:              cobra.MinimumNArgs(0),
		ValidArgsFunction: validArgsWrapper(suggestRunningServices),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				return fmt.Errorf("service name is required")
			}

			serviceName := args[0]
			if err := runShell(ctx, p, serviceName); err != nil {
				return fmt.Errorf("failed to run shell: %w", err)
			}

			return nil
		}),
	}

	root.AddCommand(cmd)
}

func runShell(ctx context.Context, p *project.Project, serviceName string) error {
	cli, err := service.New()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := cli.Shell(ctx, p, serviceName); err != nil {
		return fmt.Errorf("failed to run shell: %w", err)
	}

	return nil
}
