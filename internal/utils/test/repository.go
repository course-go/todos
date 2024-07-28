package test

import (
	"context"
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

func NewTestRepository(ctx context.Context, t *testing.T, c *postgres.PostgresContainer) *repository.Repository {
	t.Helper()
	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("failed getting container host: %v", err)
	}

	port, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("failed getting container port: %v", err)
	}

	cfg := config.Database{
		Protocol: "postgres",
		User:     dbUser,
		Password: dbPass,
		Host:     host,
		Port:     port.Port(),
		Name:     dbName,
	}
	logger := NewTestLogger(t)
	r, err := repository.New(ctx, logger, &cfg)
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

func RestoreDatabase(ctx context.Context, t *testing.T, c *postgres.PostgresContainer) {
	t.Helper()
	err := c.Restore(ctx, postgres.WithSnapshotName("test-todos"))
	if err != nil {
		t.Fatalf("failed restoring database: %v", err)
	}
}
