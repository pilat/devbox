package main

import (
	"fmt"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	var name string
	var services []string

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart services in devbox project",
		Long:  "That command will restart services in devbox project",
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

			if err := app.UpdateSources(); err != nil {
				return fmt.Errorf("failed to update sources: %w", err)
			}

			if err := app.Restart(services, true); err != nil {
				return fmt.Errorf("failed to restart services: %w", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringArrayVarP(&services, "services", "s", []string{}, "Services to restart")

	root.AddCommand(cmd)
}
