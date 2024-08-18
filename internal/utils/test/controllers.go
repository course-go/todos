package test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/course-go/todos/internal/controllers"
	"github.com/course-go/todos/internal/repository"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func NewTestRouter(ctx context.Context, t *testing.T, logger *slog.Logger) http.Handler {
	t.Helper()
	c := NewTestContainer(ctx, t)
	t.Cleanup(func() {
		err := c.Terminate(ctx)
		if err != nil {
			t.Logf("failed terminating postgres container: %v", err)
		}
	})
	cfg := NewTestDatabaseConfig(ctx, t, c)
	err := repository.Migrate(cfg, logger)
	if err != nil {
		t.Fatalf("failed migrating database: %v", err)
	}

	SeedDatabase(ctx, t, c)
	err = c.Snapshot(ctx, postgres.WithSnapshotName("test-todos"))
	if err != nil {
		t.Fatalf("failed creating database snapshot: %v", err)
	}

	r := NewTestRepository(ctx, t, logger, cfg)
	return controllers.NewAPIRouter(logger, nil, TimeNow(t), r)
}
