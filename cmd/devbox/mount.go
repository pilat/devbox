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

			if err := app.LoadProject(name); err != nil {
				return fmt.Errorf("failed to load project: %w", err)
			}

			if err := app.Mount(sourceName, targetPath); err != nil {
				return fmt.Errorf("failed to mount source code: %w", err)
			}

			if err := app.Info(); err != nil {
				return fmt.Errorf("failed to get project info: %w", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&sourceName, "source", "s", "", "Source name")
	cmd.PersistentFlags().StringVarP(&targetPath, "path", "p", "", "Path to mount")

	root.AddCommand(cmd)
}
