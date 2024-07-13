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

func NewRouter(logger *slog.Logger, config *config.Config, repository *repository.Repository) *http.ServeMux {
	mux := http.NewServeMux()
	logger = logger.With("component", "api")
	api := &API{
		logger:     logger,
		config:     config,
		repository: repository,
	}
	api.mountTodoControllers(mux)
	return mux
}

func (a API) mountTodoControllers(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/todos", a.getTodos)
	mux.HandleFunc("GET /v1/todos/{id}", a.getTodo)
	mux.HandleFunc("POST /v1/todos", a.createTodo)
	mux.HandleFunc("PUT /v1/todos/{id}", a.updateTodo)
	mux.HandleFunc("DELETE /v1/todos/{id}", a.deleteTodo)
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
