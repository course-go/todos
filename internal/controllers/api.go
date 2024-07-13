package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/todos"
)

type API struct {
	logger     *slog.Logger
	config     *config.Config
	repository *todos.Repository
}

func NewRouter(logger *slog.Logger, config *config.Config, repository *todos.Repository) *http.ServeMux {
	mux := http.NewServeMux()
	logger = logger.With("component", "api")
	api := &API{
		logger:     logger,
		config:     config,
		repository: repository,
	}
	mux.HandleFunc("GET /v1/todos", api.getTodos)
	mux.HandleFunc("GET /v1/todos/{id}", api.getTodo)
	mux.HandleFunc("POST /v1/todos", api.createTodo)
	mux.HandleFunc("PUT /v1/todos/{id}", api.updateTodo)
	mux.HandleFunc("DELETE /v1/todos/{id}", api.deleteTodo)
	return mux
}

type Response struct {
	Data  map[string]any `json:"data,omitempty"`
	Error string         `json:"error,omitempty"`
}

func responseErrorBytes(httpCode int) []byte {
	response := Response{
		Error: http.StatusText(httpCode),
	}
	bytes, _ := json.Marshal(response) // ignores error for convenience
	return bytes
}
