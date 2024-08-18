package controllers_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/course-go/todos/internal/controllers"
	"github.com/course-go/todos/internal/repository"
	ttime "github.com/course-go/todos/internal/time"
	"github.com/course-go/todos/internal/utils/test"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func newTestRouter(ctx context.Context, t *testing.T, logger *slog.Logger) http.Handler {
	t.Helper()
	c := test.NewTestContainer(ctx, t)
	t.Cleanup(func() {
		err := c.Terminate(ctx)
		if err != nil {
			t.Logf("failed terminating postgres container: %v", err)
		}
	})
	cfg := test.NewTestDatabaseConfig(ctx, t, c)
	err := repository.Migrate(cfg, logger)
	if err != nil {
		t.Fatalf("failed migrating database: %v", err)
	}

	test.SeedDatabase(ctx, t, c)
	err = c.Snapshot(ctx, postgres.WithSnapshotName("test-todos"))
	if err != nil {
		t.Fatalf("failed creating database snapshot: %v", err)
	}

	r := test.NewTestRepository(ctx, t, logger, cfg)
	return controllers.NewAPIRouter(logger, nil, newTimeNow(t), r)
}

func newTimeNow(t *testing.T) ttime.Factory {
	t.Helper()
	return func() time.Time {
		time, err := time.Parse(time.RFC3339Nano, "2024-08-18T12:14:45.847679Z")
		if err != nil {
			t.Fatalf("could not parse time: %v", err)
		}

		return time
	}
}

func TestAPIValidateSchema(t *testing.T) {
	t.Skip() // TODO: kin-openapi does not currently support OpenAPI v3.1
	ctx := context.Background()
	doc, err := openapi3.NewLoader().LoadFromFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("failed loading openapi spec from file: %v", err)
	}

	err = doc.Validate(ctx)
	if err != nil {
		t.Fatalf("failed validation openapi spec: %v", err)
	}
}
