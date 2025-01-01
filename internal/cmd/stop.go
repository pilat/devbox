package cmd

import (
	"os"

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
				os.Exit(1)
			}

			app = app.WithProject(name)

			if err := app.LoadProject(); err != nil {
				return err
			}

			return app.Stop()
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")

	return cmd
}
