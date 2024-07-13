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
