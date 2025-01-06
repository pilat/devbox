package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	var name string
	var branch string

	cmd := &cobra.Command{
		Use:   "init <git-source>",
		Short: "Initialize devbox project",
		Long:  "That command will clone devbox project from git to your ~/.devbox directory and will keep it up to date",
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
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

			if err := app.Init(name, gitURL, branch); err != nil {
				return fmt.Errorf("failed to init project: %w", err)
			}

			if err := app.LoadProject(name); err != nil {
				return fmt.Errorf("failed to load project: %w", err)
			}

			if err := app.UpdateSources(); err != nil {
				return fmt.Errorf("failed to update sources: %w", err)
			}

			if err := app.Info(); err != nil {
				return fmt.Errorf("failed to get info: %w", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Branch to clone")

	root.AddCommand(cmd)
}

func guessName(source string) string {
	elems := strings.Split(source, "/")
	name := elems[len(elems)-1]

	if strings.HasSuffix(strings.ToLower(name), ".git") {
		name = name[:len(name)-4]
	}

	return name
}
