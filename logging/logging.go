package logging

import (
	"os"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

func Setup() {
	opts := slog.HandlerOptions{
		Level: getLevel(),
	}
	handler := opts.NewTextHandler(os.Stdout)
	slog.SetDefault(slog.New(handler))
}

func getLevel() slog.Leveler {
	switch viper.GetString(config.Logging_Level) {
	case levelDebug:
		return slog.LevelDebug
	case levelInfo:
		return slog.LevelInfo
	case levelWarn:
		return slog.LevelWarn
	case levelError:
		return slog.LevelError
	}

	return slog.LevelInfo
}
