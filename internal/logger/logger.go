package logger

import (
	"errors"
	"log/slog"
	"os"

	"github.com/course-go/todos/internal/config"
)

var ErrUnknownLogLevel = errors.New("unknown log level")

func New(config *config.Logging) (logger *slog.Logger, err error) {
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, err
	}

	opts := slog.HandlerOptions{
		Level: level,
	}
	logger = slog.New(slog.NewTextHandler(os.Stdout, &opts))
	return logger, nil
}

func parseLogLevel(logLevel string) (level slog.Level, err error) {
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warm":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		err = ErrUnknownLogLevel
	}

	return level, err
}
