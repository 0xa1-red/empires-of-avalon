package main

import (
	"os"

	"golang.org/x/exp/slog"
)

func setLogging() {
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := opts.NewTextHandler(os.Stdout)
	slog.SetDefault(slog.New(handler))
}
