package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version string
	commit  string
	date    string
)

func init() {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of devbox",
		Long:  "That command will print the version number of devbox",
		Args:  cobra.MinimumNArgs(0),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			fmt.Printf("Version: %s\nCommit: %s\nCommit date: %s\n", version, commit, date)

			return nil
		}),
	}

	_ = cmd.Flags().MarkHidden("name")

	root.AddCommand(cmd)
}
