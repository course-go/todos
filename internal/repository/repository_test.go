package repository

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/todos"
	"github.com/google/uuid"
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
	err := c.Snapshot(ctx, postgres.WithSnapshotName("test-todos"))
	if err != nil {
		t.Fatalf("failed creating database snapshot: %v", err)
	}

	t.Run("Create todo", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDatabase(ctx, t, c)
		})

		r := newTestRepository(ctx, t, c)
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
			t.Fatalf("todo descriptions do not match: expected: %s != actual: %s",
				todo.Description,
				retrievedTodo.Description,
			)
		}
	})
	t.Run("Get existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDatabase(ctx, t, c)
		})

		r := newTestRepository(ctx, t, c)
		id, err := uuid.Parse("f52bad23-c201-414e-9bdb-af4327c42aa7")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		todo, err := r.GetTodo(ctx, id)
		if err != nil {
			t.Fatalf("could not get todo: %v", err)
		}

		expectedDescription := "Vacuum"
		if todo.Description != expectedDescription {
			t.Fatalf("todo descriptions do not match: expected: %s != actual: %s",
				expectedDescription,
				todo.Description,
			)
		}
	})
	t.Run("Get non-existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDatabase(ctx, t, c)
		})

		r := newTestRepository(ctx, t, c)
		id, err := uuid.Parse("be95c29a-c4dd-4d31-a5c4-d229f3374ab7")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		_, err = r.GetTodo(ctx, id)
		if !errors.Is(err, ErrTodoNotFound) {
			t.Fatalf("todo should not be found: expected: %v != actual: %v", ErrTodoNotFound, err)
		}
	})
	t.Run("Get todos", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDatabase(ctx, t, c)
		})

		r := newTestRepository(ctx, t, c)
		todos, err := r.GetTodos(ctx)
		if err != nil {
			t.Fatalf("could not get todos: %v", err)
		}

		expectedTodosLen := 2
		if len(todos) != expectedTodosLen {
			t.Fatalf("todos length does not match: expected: %d != actual: %d",
				expectedTodosLen,
				len(todos),
			)
		}
	})
	t.Run("Save existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDatabase(ctx, t, c)
		})

		r := newTestRepository(ctx, t, c)
		id, err := uuid.Parse("62446c85-3798-471f-abb8-75c1cdd7153b")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		todo, err := r.GetTodo(ctx, id)
		if err != nil {
			t.Fatalf("could not get todo: %v", err)
		}

		now := time.Now()
		todo.CompletedAt = &now
		savedTodo, err := r.SaveTodo(ctx, todo)
		if err != nil {
			t.Fatalf("could not save todo: %v", err)
		}

		if savedTodo.UpdatedAt == nil {
			t.Fatalf("todo updated timestampt was not changed")
		}

		if !savedTodo.CompletedAt.Equal(now) {
			t.Fatalf("todo completed timestampt does not match: expected: %s != actual: %s",
				now.String(),
				savedTodo.UpdatedAt.String(),
			)
		}
	})
	t.Run("Save non-existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDatabase(ctx, t, c)
		})

		r := newTestRepository(ctx, t, c)
		id, err := uuid.Parse("ac4011ce-59c9-4361-8abf-10abd273d5e5")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		todo := todos.Todo{
			ID:          id,
			Description: "Do some shopping",
			CompletedAt: nil,
		}
		_, err = r.SaveTodo(ctx, todo)
		if !errors.Is(err, ErrTodoNotFound) {
			t.Fatalf("todo should not be found: expected: %v != actual: %v", ErrTodoNotFound, err)
		}
	})
	t.Run("Delete existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDatabase(ctx, t, c)
		})

		r := newTestRepository(ctx, t, c)
		id, err := uuid.Parse("f52bad23-c201-414e-9bdb-af4327c42aa7")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		err = r.DeleteTodo(ctx, id)
		if err != nil {
			t.Fatalf("todo should be deleted: expected: nil != actual: %v", err)
		}
	})
	t.Run("Delete non-existing todo", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDatabase(ctx, t, c)
		})

		r := newTestRepository(ctx, t, c)
		id, err := uuid.Parse("4fabcaa9-7fe6-4129-86f2-1d62d142a67b")
		if err != nil {
			t.Fatalf("could not parse uuid: %v", err)
		}

		err = r.DeleteTodo(ctx, id)
		if !errors.Is(err, ErrTodoNotFound) {
			t.Fatalf("todo should not be found: expected: %v != actual: %v", ErrTodoNotFound, err)
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
		postgres.WithInitScripts(
			"migrations/20240713140024_init.up.sql",
			"testdata/seed.sql",
		),
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

func restoreDatabase(ctx context.Context, t *testing.T, c *postgres.PostgresContainer) {
	t.Helper()
	err := c.Restore(ctx, postgres.WithSnapshotName("test-todos"))
	if err != nil {
		t.Fatalf("failed restoring database: %v", err)
	}
}
