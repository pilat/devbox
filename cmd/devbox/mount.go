package main

import (
	"fmt"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	var name string
	var sourceName string
	var targetPath string

	cmd := &cobra.Command{
		Use:   "mount",
		Short: "Mount source code",
		Long:  "That command will mount source code to the project",
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

			fmt.Println("Mounting source...")
			if err := app.Mount(sourceName, targetPath); err != nil {
				return err
			}

			fmt.Println("")

			return app.Info()
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&sourceName, "source", "s", "", "Source name")
	cmd.PersistentFlags().StringVarP(&targetPath, "path", "p", "", "Path to mount")

	root.AddCommand(cmd)
}
