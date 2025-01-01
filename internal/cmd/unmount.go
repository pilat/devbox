package cmd

import (
	"os"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func NewUnmountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unmount",
		Short: "Unmount source code",
		Long:  "That command will unmount source code from the project",
		Run: func(cmd *cobra.Command, args []string) {
			app, err := app.New()
			if err != nil {
				os.Exit(1)
			}

			err = app.Unmount(name, sourceName)
			if err != nil {
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&sourceName, "source", "s", "", "Source name")

	return cmd
}
