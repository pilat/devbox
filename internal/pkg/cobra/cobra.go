package cobra

import (
	"errors"
	"strings"

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

		var errsAsStrings []string
		for {
			err = errors.Unwrap(err)
			if err == nil {
				break
			}
			errsAsStrings = append(errsAsStrings, err.Error())
		}

		var outErrors []string
		var strToRemove string
		for i := len(errsAsStrings) - 1; i >= 0; i-- {
			txt := errsAsStrings[i]
			txt = strings.ReplaceAll(txt, ": "+strToRemove, "")
			outErrors = append(outErrors, " "+txt)
			strToRemove += txt
		}

		for i, j := 0, len(outErrors)-1; i < j; i, j = i+1, j-1 {
			outErrors[i], outErrors[j] = outErrors[j], outErrors[i]
		}

		return errors.New("\n" + strings.Join(outErrors, "\n"))
	}
}
