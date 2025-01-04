package main

import (
	"fmt"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	var name string

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start devbox project",
		Long:  "That command will start devbox project",
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

			fmt.Println("Updating sources...")
			if err := app.UpdateSources(); err != nil {
				return err
			}

			fmt.Println("")

			fmt.Println("Building project...")
			if err := app.Build(); err != nil {
				return err
			}

			fmt.Println("")

			fmt.Println("Starting project...")
			return app.Up()
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")

	root.AddCommand(cmd)
}
