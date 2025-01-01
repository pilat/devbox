package main

import (
	"os"

	"github.com/pilat/devbox/internal/cmd"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use: "devbox",
	}

	root.AddCommand(
		cmd.NewInitCommand(),
		cmd.NewStartCommand(),
		cmd.NewStopCommand(),
		cmd.NewListCommand(),
		cmd.NewInfoCommand(),
		cmd.NewMountCommand(),
		cmd.NewUnmountCommand(),
	)

	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
