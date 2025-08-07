package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/course-go/todos/internal/http/controllers/health"
	"github.com/course-go/todos/internal/http/controllers/todos"
	"github.com/course-go/todos/internal/http/dto/response"
	"github.com/course-go/todos/internal/http/metrics"
	"github.com/course-go/todos/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultServerReadHeaderTimeout  = 2 * time.Second
	defaultServerIdleTimeoutTimeout = 30 * time.Second
)

func NewServer(
	logger *slog.Logger,
	metrics *metrics.Metrics,
	hostname string,
	healthController *health.HealthController,
	todosController *todos.TodosController,
) (server *http.Server, err error) {
	commonMiddleware := []middleware.Middleware{
		middleware.Logging(logger),
		middleware.Metrics(metrics),
	}
	jsonMiddleware := []middleware.Middleware{
		middleware.Logging(logger),
		middleware.Metrics(metrics),
		middleware.ContentType,
	}

	mux := chi.NewRouter()
	mux.NotFound(notFound)
	mux.MethodNotAllowed(methodNotAllowed)

	mux.With(commonMiddleware...).Handle("/metrics", promhttp.Handler())
	mux.Route("/api/v1", func(r chi.Router) {
		r.Use(jsonMiddleware...)
		r.Route("/healthz", func(r chi.Router) {
			r.Get("/", healthController.GetHealthController)
		})
		r.Route("/todos", func(r chi.Router) {
			r.Get("/", todosController.GetTodos)
			r.Get("/{id}", todosController.GetTodo)
			r.Post("/", todosController.CreateTodo)
			r.Put("/{id}", todosController.UpdateTodo)
			r.Delete("/{id}", todosController.DeleteTodo)
		})
	})

	return &http.Server{
		Addr:              hostname,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: defaultServerReadHeaderTimeout,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       defaultServerIdleTimeoutTimeout,
		Handler:           mux,
	}, nil
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	code := http.StatusNotFound
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(response.ErrorBytes(code))
}

func methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	code := http.StatusMethodNotAllowed
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(response.ErrorBytes(code))
}
