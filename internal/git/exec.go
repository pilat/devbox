package git

import (
	"context"
	"os"
	"os/exec"
)

// CommandRunner executes shell commands.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
	RunWithTTY(ctx context.Context, name string, args ...string) (string, error)
}

type defaultRunner struct{}

func (r *defaultRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (r *defaultRunner) RunWithTTY(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	return string(out), err
}
