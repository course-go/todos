package logger

import (
	"errors"
	"testing"

	"github.com/course-go/todos/internal/config"
)

func TestLogger(t *testing.T) {
	t.Parallel()
	t.Run("Valid configuration", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Logging{
			Level: "info",
		}

		_, err := New(cfg)
		if err != nil {
			t.Fatalf("could not create logger: expected: nil != actual: %v", err)
		}
	})
	t.Run("Invalid configuration", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Logging{
			Level: "what-even-is-this",
		}

		_, err := New(cfg)
		if !errors.Is(err, ErrUnknownLogLevel) {
			t.Fatalf("logger should not be created: expected: %v != actual: %v", ErrUnknownLogLevel, err)
		}
	})
}
