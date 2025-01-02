package cmd

import (
	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func NewStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop devbox project",
		Long:  "That command will stop devbox project",
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

			return app.Stop()
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")

	return cmd
}
