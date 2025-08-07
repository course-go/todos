package todos

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/course-go/todos/internal/http/dto/request"
	"github.com/course-go/todos/internal/http/dto/response"
	"github.com/course-go/todos/internal/repository"
	"github.com/course-go/todos/internal/time"
	"github.com/course-go/todos/internal/todos"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TodosController struct {
	validator  *validator.Validate
	repository *repository.Repository
	time       time.Factory
}

func NewTodosController(
	validator *validator.Validate,
	repository *repository.Repository,
	time time.Factory,
) *TodosController {
	return &TodosController{
		validator:  validator,
		repository: repository,
		time:       time,
	}
}

func (tc *TodosController) GetTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := tc.repository.GetTodos(r.Context())
	if err != nil {
		slog.Error("failed retrieving todos",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	bytes, err := response.DataBytes("todos", todos)
	if err != nil {
		slog.Error("failed constructing response",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	_, _ = w.Write(bytes)
}

func (tc *TodosController) GetTodo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		slog.Error("failed parsing uuid",
			"uuid", r.PathValue("id"),
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	todo, err := tc.repository.GetTodo(r.Context(), id)
	if errors.Is(err, repository.ErrTodoNotFound) {
		slog.Error("todo with given id does not exist",
			"error", err,
			"id", id.String(),
		)

		code := http.StatusNotFound
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	if err != nil {
		slog.Error("failed retrieving todo",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	bytes, err := response.DataBytes("todo", todo)
	if err != nil {
		slog.Error("failed constructing response",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	_, _ = w.Write(bytes)
}

func (tc *TodosController) CreateTodo(w http.ResponseWriter, r *http.Request) {
	body := r.Body

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		slog.Error("failed reading request body",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	defer func() {
		_ = body.Close()
	}()

	var req request.CreateTodoRequest

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		slog.Error("failed binding request body",
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	err = tc.validator.Struct(req)
	if err != nil {
		slog.Warn("failed validating request body",
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	todo := todos.Todo{
		Description: req.Description,
		CreatedAt:   tc.time(),
	}

	todo, err = tc.repository.CreateTodo(r.Context(), todo)
	if err != nil {
		slog.Error("failed creating todo",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	bytes, err := response.DataBytes("todo", todo)
	if err != nil {
		slog.Error("failed constructing response",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(bytes)
}

func (tc *TodosController) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		slog.Error("failed parsing uuid",
			"uuid", r.PathValue("id"),
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

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
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	defer func() {
		_ = body.Close()
	}()

	var req request.UpdateTodoRequest

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		slog.Error("failed binding request body",
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	err = tc.validator.Struct(req)
	if err != nil {
		slog.Warn("failed validating request body",
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	now := tc.time()
	todo := todos.Todo{
		ID:          id,
		Description: req.Description,
		CompletedAt: req.CompletedAt,
		UpdatedAt:   &now,
	}

	todo, err = tc.repository.SaveTodo(r.Context(), todo)
	if errors.Is(err, repository.ErrTodoNotFound) {
		slog.Warn("failed saving todo",
			"error", err,
			"id", todo.ID,
		)

		code := http.StatusNotFound
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	if err != nil {
		slog.Error("failed saving todo",
			"error", err,
			"id", todo.ID,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	bytes, err := response.DataBytes("todo", todo)
	if err != nil {
		slog.Error("failed constructing response",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bytes)
}

func (tc *TodosController) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		slog.Error("failed parsing uuid",
			"error", err,
			"uuid", r.PathValue("id"),
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	err = tc.repository.DeleteTodo(r.Context(), id, tc.time())
	if errors.Is(err, repository.ErrTodoNotFound) {
		slog.Debug("no matching id for todo",
			"id", id,
		)

		code := http.StatusNotFound
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	if err != nil {
		slog.Error("failed deleting todo",
			"error", err,
			"id", id,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
