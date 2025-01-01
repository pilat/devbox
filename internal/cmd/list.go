package cmd

import (
	"os"

	"github.com/pilat/devbox/internal/cli"
	"github.com/pilat/devbox/internal/log"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List devbox projects",
		Long:  "That command will list all devbox projects",
		Run: func(cmd *cobra.Command, args []string) {
			log := log.New()

			cli := cli.New(log)

			log.Info("List projects")
			err := cli.List()
			if err != nil {
				log.Error("Failed to list projects", "error", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
