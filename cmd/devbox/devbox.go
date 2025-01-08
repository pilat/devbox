package main

import (
	"context"
	"os"

	"github.com/pilat/devbox/internal/errors"
	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

var name string // Project name

var root = &cobra.Command{}

func main() {
	root.Use = "devbox"
	root.SetErrPrefix("Error has occurred while executing the command:\n")

	root.PersistentFlags().StringVarP(&name, "name", "n", "", "Project name")

	_ = root.RegisterFlagCompletionFunc("name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		results, err := manager.ListProjects(toComplete)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return results, cobra.ShellCompDirectiveNoFileComp
	})

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func validArgsWrapper(f func(ctx context.Context, cmd *cobra.Command, p *project.Project, args []string, toComplete string) ([]string, cobra.ShellCompDirective)) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := context.Background()
		project, err := getProject(ctx)

		if err != nil && !cmd.Flags().Changed("name") { // not auto-detected and name is not even mentioned
			return []string{"--name"}, cobra.ShellCompDirectiveNoFileComp
		} else if err != nil { // project still not found
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return f(ctx, cmd, project, args, toComplete)
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

func runWrapperWithProject(f func(ctx context.Context, p *project.Project, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		project, err := getProject(ctx)
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true

		err = f(context.Background(), project, cmd, args)

		if err == nil {
			return nil
		}

		return errors.AsStacktrace(err)
	}
}

func getProject(ctx context.Context) (*project.Project, error) {
	if name == "" {
		if detectedName, _, err := manager.Autodetect(); err == nil {
			name = detectedName
		}
	}

	return project.New(ctx, name)
}
