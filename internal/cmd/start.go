package cmd

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/fatih/color"
	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/planner"
	"github.com/spf13/cobra"
)

func NewStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start application",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Reading configuration...")
			cfg, err := config.New()
			if err != nil {
				return err
			}

			cli, err := docker.New()
			if err != nil {
				return err
			}

			defer cli.Close()

			err = cli.Ping(context.Background())
			if err != nil {
				return err
			}

			opts := &slogcolor.Options{
				Level:         slog.LevelDebug,
				TimeFormat:    time.DateTime,
				SrcFileMode:   slogcolor.ShortFile,
				SrcFileLength: 15,
				MsgPrefix:     color.HiWhiteString("| "),
			}

			log := slog.New(
				slogcolor.NewHandler(os.Stderr, opts),
			)

			ctx := context.Background()

			err = planner.Start(ctx, cli, log, cfg)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
