package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

var name string
var branch string
var sourceName string
var targetPath string

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
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := app.New()
			if err != nil {
				return err
			}

			gitURL := args[0]

			if name == "" {
				name = guessName(gitURL)
			}

			if !validateName(name) {
				return fmt.Errorf("Invalid project name: %s", name)
			}

			app = app.WithProject(name)
			return app.Init(gitURL, branch)
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Branch to clone")

	return cmd
}

func guessName(source string) string {
	elems := strings.Split(source, "/")
	name := elems[len(elems)-1]

	if strings.HasSuffix(strings.ToLower(name), ".git") {
		name = name[:len(name)-4]
	}

	return name
}

func validateName(name string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name)
}
