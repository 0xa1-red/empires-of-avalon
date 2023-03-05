package logging

import (
	"os"

	"golang.org/x/exp/slog"
)

func Setup() {
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := opts.NewTextHandler(os.Stdout)
	slog.SetDefault(slog.New(handler))
}
