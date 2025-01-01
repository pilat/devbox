package log

import (
	"log/slog"
	"os"
	"time"

	"github.com/dpotapov/slogpfx"
	"github.com/lmittmann/tint"
	"github.com/pilat/devbox/internal/pkg/utils"
)

func New() *slog.Logger {
	h := tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:  true,
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
		NoColor:    !utils.IsColorSupported(),
	})
	prefixed := slogpfx.NewHandler(h, &slogpfx.HandlerOptions{
		PrefixKeys: []string{"prefix"},
	})

	return slog.New(prefixed).With("prefix", "app")
}
