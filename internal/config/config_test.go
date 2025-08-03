package config_test

import (
	"io"
	"os"
	"testing"

	"github.com/course-go/todos/internal/config"
)

func TestConfig(t *testing.T) {
	cfgPath := "/tmp" + "/config.yaml"

	f, err := os.Create(cfgPath)
	if err != nil {
		t.Fatalf("could not create test config file: %v", err)
	}

	t.Cleanup(func() {
		err = f.Close()
		if err != nil {
			t.Errorf("could not close test config file: %v", err)
		}
	})

	_, err = io.WriteString(f,
		`database:
  protocol: postgres
  user: postgres
  password: postgres
  host: postgres
  port: 5432
  name: todos`,
	)
	if err != nil {
		t.Fatalf("could not write to test config file: %v", err)
	}

	_, err = config.Parse(cfgPath)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}
}
