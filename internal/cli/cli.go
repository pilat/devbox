package cli

import (
	"log/slog"
)

const appFolder = ".devbox"

type cli struct {
	log *slog.Logger
}

func New(log *slog.Logger) *cli {
	return &cli{
		log: log,
	}
}
