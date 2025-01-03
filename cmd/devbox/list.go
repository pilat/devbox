package main

import (
	"github.com/pilat/devbox/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List devbox projects",
		Long:  "That command will list all devbox projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := app.New()
			if err != nil {
				return err
			}

			return app.List()
		},
	}

	root.AddCommand(cmd)
}
