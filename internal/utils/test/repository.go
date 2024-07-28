package test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbUser = "todos"
	dbPass = "todos"
	dbName = "todos"
)

func NewTestRepository(
	ctx context.Context,
	t *testing.T,
	logger *slog.Logger,
	cfg *config.Database,
) *repository.Repository {
	t.Helper()
	r, err := repository.New(ctx, logger, cfg)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	return r
}

func NewTestContainer(ctx context.Context, t *testing.T) *postgres.PostgresContainer {
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

func SeedDatabase(ctx context.Context, t *testing.T, c *postgres.PostgresContainer) {
	t.Helper()
	bytes, err := os.ReadFile("testdata/seed.sql")
	if err != nil {
		t.Fatalf("could not read seed file: %v", err)
	}

	seedQuery := string(bytes)
	_, _, err = c.Exec(ctx, []string{"psql", "-U", dbUser, "-d", dbName, "-c", seedQuery})
	if err != nil {
		t.Fatalf("failed executing seeding commands: %v", err)
	}
}

func RestoreDatabase(ctx context.Context, t *testing.T, c *postgres.PostgresContainer) {
	t.Helper()
	err := c.Restore(ctx, postgres.WithSnapshotName("test-todos"))
	if err != nil {
		t.Fatalf("failed restoring database: %v", err)
	}
}
