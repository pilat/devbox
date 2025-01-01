package cmd

import (
	"os"

	"github.com/pilat/devbox/internal/cli"
	"github.com/pilat/devbox/internal/log"
	"github.com/spf13/cobra"
)

func NewInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Info devbox projects",
		Long:  "That command returns an info about a particular devbox project",
		Run: func(cmd *cobra.Command, args []string) {
			log := log.New()

			cli := cli.New(log)

			log.Info("List projects")
			err := cli.Info(name)
			if err != nil {
				log.Error("Failed to get info about project", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")

	return cmd
}
