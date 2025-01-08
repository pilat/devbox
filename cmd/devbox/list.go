package main

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/pilat/devbox/internal/table"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List devbox projects",
		Long:  "That command will list all devbox projects",
		Args:  cobra.MinimumNArgs(0),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			if err := runList(ctx, ""); err != nil {
				return fmt.Errorf("failed to list projects: %w", err)
			}

			return nil
		}),
	}

	_ = cmd.Flags().MarkHidden("name")

	root.AddCommand(cmd)
}

func runList(ctx context.Context, filter string) error {
	fmt.Println("")
	fmt.Println(" Projects:")

	projects := manager.ListProjects(filter)

	t := table.New("Name", "Message", "Author", "Date")
	for _, projectName := range projects {
		app, err := project.New(ctx, projectName)
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}

		ggg := git.New(app.WorkingDir)
		info, err := ggg.GetInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get git info: %w", err)
		}

		t.AppendRow(projectName, info.Message, info.Author, info.Date)
	}

	t.Render()

	return nil
}
