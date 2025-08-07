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
	mux := http.NewServeMux()

	// Generic controllers.
	mux.HandleFunc("/", notFound)
	mux.Handle("/metrics", promhttp.Handler())

	// Health controllers.
	mux.HandleFunc("GET /api/v1/healthz", healthController.GetHealthController)

	// Todo controllers.
	mux.HandleFunc("GET /api/v1/todos", todosController.GetTodos)
	mux.HandleFunc("GET /api/v1/todos/{id}", todosController.GetTodo)
	mux.HandleFunc("POST /api/v1/todos", todosController.CreateTodo)
	mux.HandleFunc("PUT /api/v1/todos/{id}", todosController.UpdateTodo)
	mux.HandleFunc("DELETE /api/v1/todos/{id}", todosController.DeleteTodo)

	handler := mountMiddleware(logger, metrics, mux)

	return &http.Server{
		Addr:              hostname,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: defaultServerReadHeaderTimeout,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       defaultServerIdleTimeoutTimeout,
		Handler:           handler,
	}, nil
}

func mountMiddleware(logger *slog.Logger, metrics *metrics.Metrics, router http.Handler) http.Handler {
	router = middleware.Logging(logger)(router)
	router = middleware.Metrics(metrics)(router)
	router = middleware.ContentType(router)

	return router
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	code := http.StatusNotFound
	w.WriteHeader(code)
	_, _ = w.Write(response.ErrorBytes(code))
}
