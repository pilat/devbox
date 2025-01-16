package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/pilat/devbox/internal/table"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List services in devbox project",
		Long:  "That command will list services in devbox project",
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if err := runPs(ctx, p); err != nil {
				return fmt.Errorf("failed to list services: %w", err)
			}

			return nil
		}),
	}

	root.AddCommand(cmd)
}

func runPs(ctx context.Context, p *project.Project) error {
	errCh := make(chan error)

	go func() {
		count := 0
		for {
			processTable := table.New("Age", "Name", "State", "Health")
			processTable.Compact()
			processTable.SortBy([]table.SortBy{
				{Name: "State", Mode: table.Dsc},
				{Name: "Age", Mode: table.AscAlphaNumeric},
			})

			opts := project.PsOptions{
				Project: p.Project,
				All:     true,
			}

			containers, err := apiService.Ps(ctx, p.Name, opts)
			if err != nil {
				errCh <- fmt.Errorf("failed to list services: %w", err)
			}

			if len(containers) == 0 {
				errCh <- nil
			}

			for _, container := range containers {
				uptimeDuration := time.Since(time.Unix(container.Created, 0))

				h := uptimeDuration.Hours()
				days := int(h / 24)
				uptimeDuration -= time.Duration(days) * 24 * time.Hour
				hh := uptimeDuration.Hours()
				mm := uptimeDuration.Minutes()
				ss := uptimeDuration.Seconds()

				uptimeStr := ""
				if days > 0 {
					uptimeStr += fmt.Sprintf("%dd ", days)
				} else {
					uptimeStr += fmt.Sprintf("%02d:%02d:%02d", int(hh), int(mm)%60, int(ss)%60)
				}

				if container.State == "exited" {
					uptimeStr = ""
				}

				name := container.Labels[project.ServiceLabel]
				if name == "" {
					name = container.Name
				}

				processTable.AppendRow(uptimeStr, name, container.State, container.Health)
			}

			// only return caret to top left corner
			fmt.Print("\033[H")

			// every 5 seconds, clear the screen (to reduce flickering)
			if count%20 == 0 {
				fmt.Print("\033[2J")
			}

			processTable.Render()

			time.Sleep(250 * time.Millisecond)
		}
	}()

	if err := <-errCh; err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	return nil
}
