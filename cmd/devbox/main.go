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

	root.AddCommand(cmd.NewStartCommand())
	root.AddCommand(cmd.NewStopCommand())

	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
