package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage project configuration",
		Long:  "Provides commands to manage project configuration, such as editing the environment file",
	}

	configEnvCmd := &cobra.Command{
		Use:   "env",
		Short: "Edit the environment configuration",
		Long:  "Opens the project's .env file in the default editor for editing",
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			project, err := mgr.AutodetectProject(ctx, projectName)
			if err != nil {
				return fmt.Errorf("failed to detect project: %w", err)
			}

			if err := runEditEnv(project); err != nil {
				return fmt.Errorf("failed to edit environment configuration: %w", err)
			}

			return nil
		}),
	}

	configCmd.AddCommand(configEnvCmd)
	root.AddCommand(configCmd)
}

func runEditEnv(p *project.Project) error {
	filePath := filepath.Join(p.WorkingDir, ".env")

	cmd := exec.Command(getDefaultEditor(), filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error opening file with editor: %w", err)
	}

	return nil
}

func getDefaultEditor() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}

	if editor != "" {
		return editor
	}

	for _, e := range []string{"vim", "vi", "nano"} {
		if _, err := exec.LookPath(e); err == nil {
			return e
		}
	}

	return ""
}
