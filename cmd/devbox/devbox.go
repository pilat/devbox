package main

import (
	"os"

	"github.com/pilat/devbox/internal/cobra"
)

var root = cobra.New()

func main() {
	root.Use = "devbox"
	root.SetErrPrefix("Error has occurred while executing the command:")

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
