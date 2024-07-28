package test

import (
	"context"
	"testing"

	"github.com/course-go/todos/internal/config"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func NewTestDatabaseConfig(ctx context.Context, t *testing.T, c *postgres.PostgresContainer) *config.Database {
	t.Helper()
	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("failed getting container host: %v", err)
	}

	port, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("failed getting container port: %v", err)
	}

	return &config.Database{
		Protocol: "postgres",
		User:     dbUser,
		Password: dbPass,
		Host:     host,
		Port:     port.Port(),
		Name:     dbName,
	}
}
