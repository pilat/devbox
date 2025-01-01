package cmd

import (
	"os"

	"github.com/pilat/devbox/internal/cli"
	"github.com/pilat/devbox/internal/log"
	"github.com/spf13/cobra"
)

var name string
var branch string

func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <git-source>",
		Short: "Initialize devbox project",
		Long:  "That command will clone devbox project from git to your ~/.devbox directory and will keep it up to date",
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			log := log.New()

			cli := cli.New(log)

			gitURL := args[0]

			log.Info("Init project")
			err := cli.Init(gitURL, name, branch)
			if err != nil {
				log.Error("Failed to initialize project", "error", err)
				os.Exit(1)
			}

			return
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Branch to clone")

	return cmd
}
