package main

import (
	"fmt"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	var name string
	var sourceName string

	cmd := &cobra.Command{
		Use:   "unmount",
		Short: "Unmount source code",
		Long:  "That command will unmount source code from the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := app.New()
			if err != nil {
				return err
			}

			if err := app.WithProject(name); err != nil {
				return err
			}

			if err := app.LoadProject(); err != nil {
				return err
			}

			fmt.Println("Unmounting source...")
			if err := app.Unmount(sourceName); err != nil {
				return err
			}

			fmt.Println("")

			return app.Info()
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&sourceName, "source", "s", "", "Source name")

	root.AddCommand(cmd)
}
