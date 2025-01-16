package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show logs of services in devbox project",
		Long:  "That command will show logs of services in devbox project",
		Args:  cobra.MinimumNArgs(0),
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			results, err := getRunningServices(ctx, apiService, p, true, toComplete)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			return results, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if err := runLogs(ctx, p, args); err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}

			return nil
		}),
	}

	root.AddCommand(cmd)
}

func runLogs(ctx context.Context, p *project.Project, services []string) error {
	opts := project.LogOptions{
		Project:  p.Project,
		Services: services,
		Tail:     "500",
		Follow:   true,
	}

	outStream := streams.NewOut(os.Stdout)
	errStream := streams.NewOut(os.Stderr)

	consumer := formatter.NewLogConsumer(ctx, outStream, errStream, true, true, false)
	if err := apiService.Logs(ctx, p.Name, consumer, opts); err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}
