package main

import (
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

			if err := app.WithProject(name); err != nil {
				return err
			}

			if err := app.UpdateProject(); err != nil {
				return err
			}

			if err := app.LoadProject(); err != nil {
				return err
			}

			return app.Info()
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")

	root.AddCommand(cmd)
}
