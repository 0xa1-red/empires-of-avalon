package logging

import (
	"io"
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

var fileWriter *os.File

func Setup() error {
	opts := slog.HandlerOptions{ // nolint
		Level: getLevel(),
	}

	writers := []io.Writer{os.Stdout}

	if path := viper.GetString(config.Logging_Path); path != "" {
		fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}

		fileWriter = fp

		writers = append(writers, fileWriter)
	}

	w := io.MultiWriter(writers...)

	handler := opts.NewTextHandler(w)
	slog.SetDefault(slog.New(handler))

	return nil
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

func Close() error {
	if fileWriter != nil {
		return fileWriter.Close()
	}

	return nil
}
