package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/compose/v5/pkg/api"
	"github.com/spf13/cobra"

	"github.com/pilat/devbox/internal/project"
)

func init() {
	root.AddCommand(newRunCmd())
}

func newRunCmd() *cobra.Command {
	var noTty bool

	cmd := &cobra.Command{
		Use:   "run <scenario>",
		Short: "Run scenario defined in devbox project",
		Long:  "You can pass additional arguments to the scenario",
		Args:  cobra.MinimumNArgs(1),
		ValidArgsFunction: validArgsWrapper(
			func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				p, err := mgr.AutodetectProject(ctx, projectName)
				if err != nil {
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				}

				// If a scenario is already provided (or project is not detected), disallow further completions
				if len(args) > 0 || p == nil {
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				}

				return p.GetScenarios(toComplete), cobra.ShellCompDirectiveNoFileComp
			},
		),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := mgr.AutodetectProject(ctx, projectName)
			if err != nil {
				return fmt.Errorf("failed to detect project: %w", err)
			}

			scenario, passthrough := splitScenarioArgs(args)

			if err := runRun(ctx, p, scenario, passthrough, noTty); err != nil {
				return fmt.Errorf("failed to run scenario: %w", err)
			}

			return nil
		}),
	}

	cmd.Flags().BoolVarP(&noTty, "no-tty", "t", false, "Do not allocate a pseudo-TTY")
	cmd.Flags().SetInterspersed(false)

	return cmd
}

// splitScenarioArgs returns the scenario name and its forwarded args, consuming one leading "--".
func splitScenarioArgs(args []string) (scenario string, passthrough []string) {
	scenario = args[0]
	passthrough = args[1:]

	// Interspersing is off, so pflag never consumes the "--"; drop one leading
	// separator so `run e2e -- --tag` and `run e2e --tag` forward the same args.
	if len(passthrough) > 0 && passthrough[0] == "--" {
		passthrough = passthrough[1:]
	}

	return scenario, passthrough
}

func runRun(ctx context.Context, p *project.Project, command string, args []string, noTtyFlag bool) error {
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

	commands := make([]string, 0, len(scenario.Command)+len(args))
	commands = append(commands, scenario.Command...)
	commands = append(commands, args...)

	interactive := true
	if scenario.Interactive != nil {
		interactive = *scenario.Interactive
	}

	var tty bool
	switch {
	case noTtyFlag:
		tty = false
	case scenario.Tty != nil:
		tty = *scenario.Tty
	default:
		tty = isTTYAvailable(os.Stdin)
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

func isRunning(ctx context.Context, a api.Compose, p *project.Project) (bool, error) {
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
