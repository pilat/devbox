package main

import (
	"fmt"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	var name string

	cmd := &cobra.Command{
		Use:   "run <scenario>",
		Short: "Run scenario defined in devbox project",
		Long:  "You can pass additional arguments to the scenario",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			remainingArgs := cmd.Flags().Args()

			app, err := app.New()
			if err != nil {
				return err
			}

			if err := app.LoadProject(name); err != nil {
				return fmt.Errorf("failed to load project: %w", err)
			}

			command := args[0]
			if len(args) > 1 {
				remainingArgs = args[1:]
			} else {
				remainingArgs = []string{}
			}

			if err := app.Run(command, remainingArgs); err != nil {
				return fmt.Errorf("failed to run scenario: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().SetInterspersed(false)
	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")

	root.AddCommand(cmd)
}
