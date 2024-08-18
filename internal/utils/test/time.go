package test

import (
	"testing"
	"time"

	ttime "github.com/course-go/todos/internal/time"
)

func TimeNow(t *testing.T) ttime.Factory {
	t.Helper()
	return func() time.Time {
		time, err := time.Parse(time.RFC3339Nano, "2024-08-18T12:14:45.847679Z")
		if err != nil {
			t.Fatalf("could not parse time: %v", err)
		}

		return time
	}
}
