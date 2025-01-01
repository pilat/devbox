package cmd

import (
	"os"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func NewMountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mount",
		Short: "Mount source code",
		Long:  "That command will mount source code to the project",
		Run: func(cmd *cobra.Command, args []string) {
			app, err := app.New()
			if err != nil {
				os.Exit(1)
			}

			err = app.Mount(name, sourceName, targetPath)
			if err != nil {
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&sourceName, "source", "s", "", "Source name")
	cmd.PersistentFlags().StringVarP(&targetPath, "path", "p", "", "Path to mount")

	return cmd
}
