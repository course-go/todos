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

type Controller struct {
	logger     *slog.Logger
	validator  *validator.Validate
	repository *repository.Repository
	time       time.Factory
}

func NewController(
	logger *slog.Logger,
	validator *validator.Validate,
	repository *repository.Repository,
	time time.Factory,
) *Controller {
	return &Controller{
		logger:     logger.With("component", "http.controllers.todos"),
		validator:  validator,
		repository: repository,
		time:       time,
	}
}

func (c *Controller) GetTodosController(w http.ResponseWriter, r *http.Request) {
	todos, err := c.repository.GetTodos(r.Context())
	if err != nil {
		c.logger.Error("failed retrieving todos",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	bytes, err := response.DataBytes("todos", todos)
	if err != nil {
		c.logger.Error("failed constructing response",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	_, _ = w.Write(bytes)
}

func (c *Controller) GetTodoController(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		c.logger.Error("failed parsing uuid",
			"uuid", r.PathValue("id"),
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	todo, err := c.repository.GetTodo(r.Context(), id)
	if errors.Is(err, repository.ErrTodoNotFound) {
		c.logger.Error("todo with given id does not exist",
			"error", err,
			"id", id.String(),
		)

		code := http.StatusNotFound
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	if err != nil {
		c.logger.Error("failed retrieving todo",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	bytes, err := response.DataBytes("todo", todo)
	if err != nil {
		c.logger.Error("failed constructing response",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	_, _ = w.Write(bytes)
}

func (c *Controller) CreateTodoController(w http.ResponseWriter, r *http.Request) {
	body := r.Body

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		c.logger.Error("failed reading request body",
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
		c.logger.Error("failed binding request body",
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	err = c.validator.Struct(req)
	if err != nil {
		c.logger.Warn("failed validating request body",
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	todo := todos.Todo{
		Description: req.Description,
		CreatedAt:   c.time(),
	}

	todo, err = c.repository.CreateTodo(r.Context(), todo)
	if err != nil {
		c.logger.Error("failed creating todo",
			"error", err,
		)

		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	bytes, err := response.DataBytes("todo", todo)
	if err != nil {
		c.logger.Error("failed constructing response",
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

func (c *Controller) UpdateTodoController(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		c.logger.Error("failed parsing uuid",
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
		c.logger.Error("failed reading request body",
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
		c.logger.Error("failed binding request body",
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	err = c.validator.Struct(req)
	if err != nil {
		c.logger.Warn("failed validating request body",
			"error", err,
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	now := c.time()
	todo := todos.Todo{
		ID:          id,
		Description: req.Description,
		CompletedAt: req.CompletedAt,
		UpdatedAt:   &now,
	}

	todo, err = c.repository.SaveTodo(r.Context(), todo)
	if errors.Is(err, repository.ErrTodoNotFound) {
		c.logger.Warn("failed saving todo",
			"error", err,
			"id", todo.ID,
		)

		code := http.StatusNotFound
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	if err != nil {
		c.logger.Error("failed saving todo",
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
		c.logger.Error("failed constructing response",
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

func (c *Controller) DeleteTodoController(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		c.logger.Error("failed parsing uuid",
			"error", err,
			"uuid", r.PathValue("id"),
		)

		code := http.StatusBadRequest
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	err = c.repository.DeleteTodo(r.Context(), id, c.time())
	if errors.Is(err, repository.ErrTodoNotFound) {
		c.logger.Debug("no matching id for todo",
			"id", id,
		)

		code := http.StatusNotFound
		w.WriteHeader(code)
		_, _ = w.Write(response.ErrorBytes(code))

		return
	}

	if err != nil {
		c.logger.Error("failed deleting todo",
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
