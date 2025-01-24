package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pilat/devbox/internal/manager"
	"github.com/spf13/cobra"
)

func init() {
	var projectName string
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
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			gitURL := args[0]

			if projectName == "" {
				projectName = guessName(gitURL)
			}

			if err := runInit(projectName, gitURL, branch); err != nil {
				return fmt.Errorf("failed to list projects: %w", err)
			}

			return nil
		}),
	}

	cmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch to clone")

	root.AddCommand(cmd)
}

func runInit(name, gitURL, branch string) error {
	fmt.Println("[*] Initializing project...")
	if err := manager.Init(name, gitURL, branch); err != nil {
		return fmt.Errorf("failed to init project: %w", err)
	}

	fmt.Println("\nProject has been successfully initialized!")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Configure environment (optional):\n")
	fmt.Printf("     devbox --name %s config env\n\n", name)
	fmt.Printf("  2. Start services:\n")
	fmt.Printf("     devbox --name %s up\n", name)

	return nil
}

func guessName(source string) string {
	elems := strings.Split(source, "/")
	name := elems[len(elems)-1]

	if strings.HasSuffix(strings.ToLower(name), ".git") {
		name = name[:len(name)-4]
	}

	return name
}
