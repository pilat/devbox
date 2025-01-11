package main

import (
	"context"
	"os"

	"github.com/pilat/devbox/internal/errors"
	"github.com/pilat/devbox/internal/manager"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{}

var projectName string

func main() {
	root.Use = "devbox"
	root.SetErrPrefix("Error has occurred while executing the command:\n")

	root.PersistentFlags().StringVarP(&projectName, "name", "n", "", "Project name")

	_ = root.RegisterFlagCompletionFunc("name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return manager.ListProjects(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func validArgsWrapper(f func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := context.Background()

		_, err := manager.AutodetectProject(projectName)

		if err != nil && !cmd.Flags().Changed("name") { // not auto-detected and name is not even mentioned
			return []string{"--name"}, cobra.ShellCompDirectiveNoFileComp
		} else if err != nil { // project still not found
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return f(ctx, cmd, args, toComplete)
	}
}

func runWrapper(f func(ctx context.Context, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cmd.SilenceUsage = true

		err := f(ctx, cmd, args)

		if err == nil {
			return nil
		}

		return errors.AsStacktrace(err)
	}
}
