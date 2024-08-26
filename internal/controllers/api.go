package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/controllers/metrics"
	"github.com/course-go/todos/internal/controllers/middleware"
	"github.com/course-go/todos/internal/repository"
	"github.com/course-go/todos/internal/time"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

type API struct {
	logger     *slog.Logger
	config     *config.Config
	time       time.Factory
	validator  *validator.Validate
	repository *repository.Repository
}

func NewAPIRouter(
	logger *slog.Logger,
	config *config.Config,
	provider *metric.MeterProvider,
	time time.Factory,
	repository *repository.Repository,
) (router http.Handler, err error) {
	mux := http.NewServeMux()
	logger = logger.With("component", "api")
	v := validator.New(validator.WithRequiredStructEnabled())
	api := API{
		logger:     logger,
		config:     config,
		time:       time,
		validator:  v,
		repository: repository,
	}
	metrics, err := metrics.New(provider)
	if err != nil {
		return nil, err
	}

	api.mountCommonControllers(mux)
	api.mountTodoControllers(mux)
	router = api.addMiddleware(mux, logger, metrics)
	return router, nil
}

func (a API) mountCommonControllers(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
}

func (a API) mountTodoControllers(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/todos", a.GetTodos)
	mux.HandleFunc("GET /api/v1/todos/{id}", a.GetTodo)
	mux.HandleFunc("POST /api/v1/todos", a.CreateTodo)
	mux.HandleFunc("PUT /api/v1/todos/{id}", a.UpdateTodo)
	mux.HandleFunc("DELETE /api/v1/todos/{id}", a.DeleteTodo)
}

func (a API) addMiddleware(mux *http.ServeMux, logger *slog.Logger, metrics *metrics.Metrics) http.Handler {
	loggingMiddleware := middleware.Logging(logger)
	router := loggingMiddleware(mux)
	metricsMiddleware := middleware.Metrics(metrics)
	router = metricsMiddleware(mux)
	router = middleware.ContentType(router)
	return router
}

type Response struct {
	Data  map[string]any `json:"data,omitempty"`
	Error string         `json:"error,omitempty"`
}

func responseErrorBytes(httpCode int) []byte {
	response := Response{
		Error: http.StatusText(httpCode),
	}
	bytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return bytes
}

func responseDataBytes(name string, data any) (bytes []byte, err error) {
	response := Response{
		Data: map[string]any{
			name: data,
		},
	}

	return json.Marshal(response) //nolint: wrapcheck
}
