package main

import (
	"context"
	"fmt"

	"github.com/docker/compose/v2/pkg/api"
	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "run <scenario>",
		Short: "Run scenario defined in devbox project",
		Long:  "You can pass additional arguments to the scenario",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			// If a scenario is already provided (or project is not detected), disallow further completions
			if len(args) > 0 || p == nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			return p.GetScenarios(toComplete), cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			command := args[0]
			if len(args) > 1 {
				args = args[1:]
			} else {
				args = []string{}
			}

			if err := runRun(ctx, p, command, args); err != nil {
				return fmt.Errorf("failed to run scenario: %w", err)
			}

			return nil
		}),
	}

	cmd.Flags().SetInterspersed(false)

	root.AddCommand(cmd)
}

func runRun(ctx context.Context, p *project.Project, command string, args []string) error {
	isRunning, err := isRunning(ctx, apiService, p)
	if err != nil {
		return fmt.Errorf("failed to check if services are running: %w", err)
	}

	if !isRunning {
		return nil
	}

	scenario, ok := p.Scenarios[command]
	if !ok {
		return fmt.Errorf("scenario %q not found", command)
	}

	commands := []string{}
	commands = append(commands, scenario.Command...)
	commands = append(commands, args...)

	interactive := true
	if scenario.Interactive != nil {
		interactive = *scenario.Interactive
	}

	tty := true
	if scenario.Tty != nil {
		tty = *scenario.Tty
	}

	opts := project.RunOptions{
		Service:     scenario.Service,
		Interactive: interactive,
		Tty:         tty,
		Command:     commands,
		Entrypoint:  scenario.Entrypoint,
		WorkingDir:  scenario.WorkingDir,
		User:        scenario.User,
	}

	exitCode, err := apiService.Exec(ctx, p.Name, opts)
	if err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	if exitCode != 0 {
		return fmt.Errorf("non-zero exit code: %d", exitCode)
	}

	return nil
}

func isRunning(ctx context.Context, a api.Service, p *project.Project) (bool, error) {
	opts := project.PsOptions{
		Project: p.Project,
	}

	containers, err := a.Ps(ctx, p.Name, opts)
	if err != nil {
		return false, fmt.Errorf("failed to get services: %w", err)
	}

	hasAny := false
	for _, container := range containers {
		hasAny = container.Labels[project.ProjectLabel] == p.Name &&
			container.Labels[project.WorkingDirLabel] == p.WorkingDir
		if hasAny {
			break
		}
	}

	return hasAny, nil
}
