package cobra

import (
	"github.com/pilat/devbox/internal/errors"
	"github.com/spf13/cobra"
)

type Command struct {
	*cobra.Command
}

func New() *Command {
	return &Command{
		Command: &cobra.Command{},
	}
}

func (c *Command) AddCommand(cmd *cobra.Command) {
	cmd.RunE = c.runWrapper(cmd.RunE)
	c.Command.AddCommand(cmd)
}

func (c *Command) runWrapper(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		err := f(cmd, args)

		if err == nil {
			return nil
		}

		return errors.AsStacktrace(err)
	}
}
