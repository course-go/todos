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

func (a API) GetTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := a.repository.GetTodos(r.Context())
	if err != nil {
		slog.Error("failed retrieving todos",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	bytes, err := responseDataBytes("todos", todos)
	if err != nil {
		slog.Error("failed constructing response",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.Write(bytes)
}

func (a API) GetTodo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		slog.Error("failed parsing uuid",
			"uuid", r.PathValue("id"),
			"error", err,
		)
		code := http.StatusBadRequest
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	todo, err := a.repository.GetTodo(r.Context(), id)
	if errors.Is(err, repository.ErrTodoNotFound) {
		slog.Error("todo with given id does not exist",
			"error", err,
			"id", id.String(),
		)
		code := http.StatusNotFound
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	if err != nil {
		slog.Error("failed retrieving todo",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	bytes, err := responseDataBytes("todo", todo)
	if err != nil {
		slog.Error("failed constructing response",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.Write(bytes)
}

func (a API) CreateTodo(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		slog.Error("failed reading request body",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	defer body.Close()
	var request CreateTodoRequest
	err = json.Unmarshal(bodyBytes, &request)
	if err != nil {
		slog.Error("failed binding request body",
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
	todo, err = a.repository.CreateTodo(r.Context(), todo)
	if err != nil {
		slog.Error("failed creating todo",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	bytes, err := responseDataBytes("todo", todo)
	if err != nil {
		slog.Error("failed constructing response",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(bytes)
}

func (a API) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		slog.Error("failed parsing uuid",
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
		slog.Error("failed reading request body",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	defer body.Close()
	var request UpdateTodoRequest
	err = json.Unmarshal(bodyBytes, &request)
	if err != nil {
		slog.Error("failed binding request body",
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
	todo, err = a.repository.SaveTodo(r.Context(), todo)
	if err != nil {
		slog.Error("failed saving todo",
			"error", err,
			"id", todo.ID,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	bytes, err := responseDataBytes("todo", todo)
	if err != nil {
		slog.Error("failed constructing response",
			"error", err,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func (a API) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		slog.Error("failed parsing uuid",
			"error", err,
			"uuid", r.PathValue("id"),
		)
		code := http.StatusBadRequest
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	err = a.repository.DeleteTodo(r.Context(), id)
	if errors.Is(err, repository.ErrTodoNotFound) {
		slog.Debug("no matching id for todo",
			"id", id,
		)
		code := http.StatusNotFound
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	if err != nil {
		slog.Error("failed deleting todo",
			"error", err,
			"id", id,
		)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		w.Write(responseErrorBytes(code))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
