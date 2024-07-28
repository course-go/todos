package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/repository"
)

type API struct {
	logger     *slog.Logger
	config     *config.Config
	repository *repository.Repository
}

func NewAPI(logger *slog.Logger, config *config.Config, repository *repository.Repository) API {
	return API{
		logger:     logger,
		config:     config,
		repository: repository,
	}
}

func NewRouter(logger *slog.Logger, config *config.Config, repository *repository.Repository) *http.ServeMux {
	mux := http.NewServeMux()
	logger = logger.With("component", "api")
	api := NewAPI(logger, config, repository)
	api.mountTodoControllers(mux)
	return mux
}

func (a API) mountTodoControllers(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/todos", a.GetTodos)
	mux.HandleFunc("GET /api/v1/todos/{id}", a.GetTodo)
	mux.HandleFunc("POST /api/v1/todos", a.CreateTodo)
	mux.HandleFunc("PUT /api/v1/todos/{id}", a.UpdateTodo)
	mux.HandleFunc("DELETE /api/v1/todos/{id}", a.DeleteTodo)
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
