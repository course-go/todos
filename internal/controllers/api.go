package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/controllers/middleware"
	"github.com/course-go/todos/internal/repository"
	"github.com/course-go/todos/internal/time"
)

type API struct {
	logger     *slog.Logger
	config     *config.Config
	time       time.Factory
	repository *repository.Repository
}

func NewAPIRouter(
	logger *slog.Logger,
	config *config.Config,
	time time.Factory,
	repository *repository.Repository,
) http.Handler {
	mux := http.NewServeMux()
	logger = logger.With("component", "api")
	api := API{
		logger:     logger,
		config:     config,
		time:       time,
		repository: repository,
	}
	api.mountTodoControllers(mux)
	router := api.addMiddleware(mux, logger)
	return router
}

func (a API) mountTodoControllers(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/todos", a.GetTodos)
	mux.HandleFunc("GET /api/v1/todos/{id}", a.GetTodo)
	mux.HandleFunc("POST /api/v1/todos", a.CreateTodo)
	mux.HandleFunc("PUT /api/v1/todos/{id}", a.UpdateTodo)
	mux.HandleFunc("DELETE /api/v1/todos/{id}", a.DeleteTodo)
}

func (a API) addMiddleware(mux *http.ServeMux, logger *slog.Logger) http.Handler {
	loggingMiddleware := middleware.Logging(logger)
	router := loggingMiddleware(mux)
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
