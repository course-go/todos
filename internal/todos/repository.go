package todos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/course-go/todos/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrTodoNotFound = errors.New("todo with given UUID does not exist")

type Repository struct {
	logger *slog.Logger
	config *config.Config
	conn   *pgx.Conn
	mu     sync.Mutex
	todos  []Todo
}

func NewRepository(ctx context.Context, logger *slog.Logger, config *config.Config) (repository *Repository, err error) {
	databaseURL := fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		config.Database.Protocol,
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		err = fmt.Errorf("failed connecting to database: %w", err)
		return
	}

	logger = logger.With("component", "repository")
	repository = &Repository{
		logger: logger,
		config: config,
		conn:   conn,
		todos:  make([]Todo, 0),
	}
	return
}

func (r *Repository) GetTodos() (todos []Todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.todos
}

func (r *Repository) GetTodo(id uuid.UUID) (todo Todo, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, todo := range r.todos {
		if todo.ID == id {
			return todo, nil
		}
	}

	return Todo{}, ErrTodoNotFound
}

func (r *Repository) CreateTodo(todo Todo) (createdTodo Todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	todo.ID = uuid.New()
	todo.CreatedAt = time.Now()
	r.todos = append(r.todos, todo)
	return todo
}

func (r *Repository) SaveTodo(todo Todo) (savedTodo Todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	index := slices.IndexFunc(r.todos, func(t Todo) bool {
		return t.ID == todo.ID
	})
	if index == -1 {
		todo.CreatedAt = time.Now()
		r.todos = append(r.todos, todo)
		return todo
	}

	todo.CreatedAt = r.todos[index].CreatedAt
	now := time.Now()
	todo.UpdatedAt = &now
	r.todos[index] = todo
	return todo
}

func (r *Repository) DeleteTodo(id uuid.UUID) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	index := slices.IndexFunc(r.todos, func(todo Todo) bool {
		return id == todo.ID
	})
	if index == -1 {
		return ErrTodoNotFound
	}

	slice := slices.Delete(r.todos, index, index+1)
	r.todos = slice[:]
	return nil
}
