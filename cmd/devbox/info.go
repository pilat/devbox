package main

import (
	"fmt"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	var name string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Info devbox projects",
		Long:  "That command returns an info about a particular devbox project",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := app.New()
			if err != nil {
				return err
			}

			if err := app.LoadProject(name); err != nil {
				return fmt.Errorf("failed to load project: %w", err)
			}

			if err := app.UpdateProject(); err != nil {
				return fmt.Errorf("failed to update project: %w", err)
			}

			if err := app.Info(); err != nil {
				return fmt.Errorf("failed to get project info: %w", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")

	root.AddCommand(cmd)
}
