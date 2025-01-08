package service

import (
	"context"
	"fmt"

	"github.com/pilat/devbox/internal/project"
)

func (a *Service) Run(ctx context.Context, p *project.Project, command string, args []string) error {
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

	exitCode, err := a.service.Exec(ctx, p.Name, opts)
	if err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	if exitCode != 0 {
		return fmt.Errorf("non-zero exit code: %d", exitCode)
	}

	return nil
}
