package cmd

import (
	"os"

	"github.com/pilat/devbox/internal/cli"
	"github.com/pilat/devbox/internal/log"
	"github.com/spf13/cobra"
)

func NewStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start devbox project",
		Long:  "That command will start devbox project",
		Run: func(cmd *cobra.Command, args []string) {
			log := log.New()

			cli := cli.New(log)

			log.Info("Start project")
			err := cli.Start(name)
			if err != nil {
				log.Error("Failed to start project", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")

	return cmd
}
