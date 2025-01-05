package main

import (
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

			if err := app.WithProject(name); err != nil {
				return err
			}

			if err := app.UpdateProject(); err != nil {
				return err
			}

			if err := app.LoadProject(); err != nil {
				return err
			}

			if err := app.UpdateSources(); err != nil {
				return err
			}

			return app.Restart(services, true)
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringArrayVarP(&services, "services", "s", []string{}, "Services to restart")

	root.AddCommand(cmd)
}
