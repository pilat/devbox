package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
	"github.com/pilat/devbox/internal/manager"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{}

var projectName string

var dockerClient client.APIClient
var apiService api.Service

func main() {
	for _, fn := range []func() error{
		initDocker,
		initCobra,
	} {
		if err := fn(); err != nil {
			os.Exit(1)
		}
	}
}

func initCobra() error {
	root.Use = "devbox"
	root.SetErrPrefix("Error has occurred while executing the command:")

	root.PersistentFlags().StringVarP(&projectName, "name", "n", "", "Project name")

	_ = root.RegisterFlagCompletionFunc("name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return manager.ListProjects(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return root.Execute()
}

func initDocker() error {
	dockerCLI, err := command.NewDockerCli()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	cliOpts := flags.NewClientOptions()
	if err = dockerCLI.Initialize(cliOpts); err != nil {
		return fmt.Errorf("failed to initialize docker client: %w", err)
	}

	dockerClient = dockerCLI.Client()

	apiService = compose.NewComposeService(dockerCLI)

	return nil
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

		return f(ctx, cmd, args)
	}
}
