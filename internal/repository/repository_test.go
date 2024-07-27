package repository

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/todos"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbUser = "todos"
	dbPass = "todos"
	dbName = "todos"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	c := newTestContainer(ctx, t)
	t.Cleanup(func() {
		err := c.Terminate(ctx)
		if err != nil {
			t.Logf("failed terminating postgres container: %v", err)
		}
	})
	r := newTestRepository(ctx, t, c)
	seedDatabase(ctx, t)
	err := c.Snapshot(ctx, postgres.WithSnapshotName("test-todos"))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Create todo", func(t *testing.T) {
		t.Cleanup(func() {
			err = c.Restore(ctx)
			if err != nil {
				t.Fatalf("failed restoring database: %v", err)
			}
		})

		todo := todos.Todo{
			Description: "Mop the floor",
		}
		createdTodo, err := r.CreateTodo(ctx, todo)
		if err != nil {
			t.Fatalf("could not create todo: %v", err)
		}

		retrievedTodo, err := r.GetTodo(ctx, createdTodo.ID)
		if err != nil {
			t.Fatalf("could not retrieve created todo: %v", err)
		}

		if todo.Description != retrievedTodo.Description {
			t.Fatalf("todo descriptions do not match: expected: %s != actual: %s]",
				todo.Description,
				retrievedTodo.Description,
			)
		}
	})
}

func newTestContainer(ctx context.Context, t *testing.T) *postgres.PostgresContainer {
	t.Helper()
	c, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		postgres.WithSQLDriver("pgx5"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	return c
}

func newTestRepository(ctx context.Context, t *testing.T, c *postgres.PostgresContainer) *Repository {
	t.Helper()
	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("failed getting container host: %v", err)
	}

	port, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("failed getting container port: %v", err)
	}

	cfg := config.Config{
		Database: config.Database{
			Protocol: "postgres",
			User:     dbUser,
			Password: dbPass,
			Host:     host,
			Port:     port.Port(),
			Name:     dbName,
		},
	}
	logger := newTestLogger(t)
	r, err := New(ctx, logger, &cfg)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	return r
}

func newTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &opts))
}

func seedDatabase(ctx context.Context, t *testing.T) {}
