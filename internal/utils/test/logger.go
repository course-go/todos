package test

import (
	"log/slog"
	"os"
	"testing"
)

func NewTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &opts))
}
