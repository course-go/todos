package repository

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/todos"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

var (
	ErrMigrations   = errors.New("failed migrating database schema")
	ErrTodoNotFound = errors.New("todo with given UUID does not exist")
)

type Repository struct {
	logger *slog.Logger
	config *config.Config
	pool   *pgxpool.Pool
	mu     sync.Mutex
	todos  []todos.Todo
}

func New(
	ctx context.Context,
	logger *slog.Logger,
	config *config.Config,
) (repository *Repository, err error) {
	logger = logger.With("component", "repository")
	err = migrateRepository(logger, config)
	if err != nil {
		return
	}

	databaseURL := fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		config.Database.Protocol,
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		err = fmt.Errorf("failed creating pgx pool: %w", err)
		return
	}

	repository = &Repository{
		logger: logger,
		config: config,
		pool:   pool,
		todos:  make([]todos.Todo, 0),
	}
	return
}

func (r *Repository) GetTodos() (todos []todos.Todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.todos
}

func (r *Repository) GetTodo(id uuid.UUID) (todo todos.Todo, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, todo := range r.todos {
		if todo.ID == id {
			return todo, nil
		}
	}

	return todos.Todo{}, ErrTodoNotFound
}

func (r *Repository) CreateTodo(todo todos.Todo) (createdTodo todos.Todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	todo.ID = uuid.New()
	todo.CreatedAt = time.Now()
	r.todos = append(r.todos, todo)
	return todo
}

func (r *Repository) SaveTodo(todo todos.Todo) (savedTodo todos.Todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	index := slices.IndexFunc(r.todos, func(t todos.Todo) bool {
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
	index := slices.IndexFunc(r.todos, func(todo todos.Todo) bool {
		return id == todo.ID
	})
	if index == -1 {
		return ErrTodoNotFound
	}

	slice := slices.Delete(r.todos, index, index+1)
	r.todos = slice[:]
	return nil
}

func migrateRepository(logger *slog.Logger, config *config.Config) error {
	databaseURL := fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		"pgx5", // golang-migrate uses "stdlib registered" drivers set by imports
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)
	d, err := iofs.New(embedMigrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed initializing driver from iofs: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		return fmt.Errorf("failed creating migrations: %w", err)
	}

	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			logger.Warn("failed closing migrations source: %w", srcErr)
		}

		if dbErr != nil {
			logger.Warn("failed closing database after migrations: %w", dbErr)
		}
	}()
	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		logger.Info("database schema is up to date")
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed applying migrations: %w", err)
	}

	return nil
}
