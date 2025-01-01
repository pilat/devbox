package cmd

import (
	"os"

	"github.com/pilat/devbox/internal/cli"
	"github.com/pilat/devbox/internal/log"
	"github.com/spf13/cobra"
)

func NewMountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mount",
		Short: "Mount source code",
		Long:  "That command will mount source code to the project",
		Run: func(cmd *cobra.Command, args []string) {
			log := log.New()

			cli := cli.New(log)

			log.Info("Unmount source code")
			err := cli.Mount(name, sourceName, targetPath)
			if err != nil {
				log.Error("Failed to unmount source code", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&sourceName, "source", "s", "", "Source name")
	cmd.PersistentFlags().StringVarP(&targetPath, "path", "p", "", "Path to mount")

	return cmd
}
