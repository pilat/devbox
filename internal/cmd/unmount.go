package cmd

import (
	"os"

	"github.com/pilat/devbox/internal/cli"
	"github.com/pilat/devbox/internal/log"
	"github.com/spf13/cobra"
)

func NewUnmountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unmount",
		Short: "Unmount source code",
		Long:  "That command will unmount source code from the project",
		Run: func(cmd *cobra.Command, args []string) {
			log := log.New()

			cli := cli.New(log)

			log.Info("Unmount source code")
			err := cli.Unmount(name, sourceName)
			if err != nil {
				log.Error("Failed to unmount source code", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&sourceName, "source", "s", "", "Source name")

	return cmd
}
