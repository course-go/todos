package test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/course-go/todos/internal/health"
	thttp "github.com/course-go/todos/internal/http"
	chealth "github.com/course-go/todos/internal/http/controllers/health"
	ctodos "github.com/course-go/todos/internal/http/controllers/todos"
	"github.com/course-go/todos/internal/http/metrics"
	"github.com/course-go/todos/internal/repository"
	"github.com/go-playground/validator/v10"
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
	p := NewMetricProvider(t)

	m, err := metrics.New(p)
	if err != nil {
		t.Fatalf("failed creating http metrics: %v", err)
	}

	h, err := health.NewRegistry(ctx)
	if err != nil {
		t.Fatalf("failed creating health registry: %v", err)
	}

	tc := ctodos.NewTodosController(validator.New(validator.WithRequiredStructEnabled()), r, NewTimeNow(t))
	hc := chealth.NewHealthController(h)

	server, err := thttp.NewServer(logger, m, "testing", hc, tc)
	if err != nil {
		t.Fatalf("failed creating http server: %v", err)
	}

	return server.Handler
}
