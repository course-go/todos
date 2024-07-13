package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/course-go/todos/internal/repository"
	"github.com/course-go/todos/internal/todos"
	"github.com/google/uuid"
)

type CreateTodoRequest struct {
	Description string `binding:"required" json:"description"`
}

type UpdateTodoRequest struct {
	Description string     `binding:"required" json:"description"`
	CompletedAt *time.Time `json:"completedAt"`
}

func (a API) getTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := a.repository.GetTodos(r.Context())
	if err != nil {
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	bytes, err := responseDataBytes("todos", todos)
	if err != nil {
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.Write(bytes)
}

func (a API) getTodo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		slog.Error("could not parse uuid",
			"uuid", r.PathValue("id"),
			"error", err,
		)
		code := http.StatusBadRequest
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	todo, err := a.repository.GetTodo(id)
	if errors.Is(err, repository.ErrTodoNotFound) {
		slog.Error("todo with given uuid does not exist",
			"uuid", id.String(),
			"error", err,
		)
		code := http.StatusNotFound
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	if err != nil {
		slog.Error("could not retrieve todos from database",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	bytes, err := responseDataBytes("todo", todo)
	if err != nil {
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.Write(bytes)
}

func (a API) createTodo(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	defer body.Close()
	var request CreateTodoRequest
	err = json.Unmarshal(bodyBytes, &request)
	if err != nil {
		slog.Error("unbindable body received",
			"error", err,
		)
		code := http.StatusBadRequest
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	todo := todos.Todo{
		Description: request.Description,
	}
	todo = a.repository.CreateTodo(todo)
	bytes, err := responseDataBytes("todo", todo)
	if err != nil {
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(bytes)
}

func (a API) updateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		slog.Error("could not parse uuid",
			"uuid", r.PathValue("id"),
			"error", err,
		)
		code := http.StatusBadRequest
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	body := r.Body
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	defer body.Close()
	var request UpdateTodoRequest
	err = json.Unmarshal(bodyBytes, &request)
	if err != nil {
		slog.Error("unbindable body received",
			"error", err,
		)
		code := http.StatusBadRequest
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	todo := todos.Todo{
		ID:          id,
		Description: request.Description,
		CompletedAt: request.CompletedAt,
	}
	todo = a.repository.SaveTodo(todo)
	bytes, err := responseDataBytes("todo", todo)
	if err != nil {
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func (a API) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		slog.Error("could not parse uuid",
			"uuid", r.PathValue("id"),
			"error", err,
		)
		code := http.StatusBadRequest
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	err = a.repository.DeleteTodo(id)
	if errors.Is(err, repository.ErrTodoNotFound) {
		code := http.StatusNotFound
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	if err != nil {
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
