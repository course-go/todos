package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/health"
	"github.com/course-go/todos/internal/todos"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // Used to register "pgx5" driver used for migrations.
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	databaseHealthPingPeriod = 30 * time.Second
)

var (
	ErrMigrations   = errors.New("failed migrating database schema")
	ErrTodoNotFound = errors.New("todo with given UUID does not exist")
	ErrDatabase     = errors.New("failed querying database")
)

type Repository struct {
	logger   *slog.Logger
	registry *health.Registry
	config   *config.Database
	pool     *pgxpool.Pool
}

func New(
	ctx context.Context,
	logger *slog.Logger,
	registry *health.Registry,
	config *config.Database,
) (repository *Repository, err error) {
	logger = logger.With("component", "repository")
	databaseURL := fmt.Sprintf("%s://%s:%s@%s:%s/%s?%s",
		config.Protocol,
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
		config.Options,
	)

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed creating pgx pool: %w", err)
	}

	logger.Info(
		"created pgx pool",
		"databaseURL", databaseURL,
	)

	checks := []health.Check{
		{
			Period: databaseHealthPingPeriod,
			CheckFn: func(ctx context.Context, c *health.Component) {
				c.UpdatedAt = time.Now()
				err := pool.Ping(ctx)
				if err != nil {
					c.Health = health.ERROR
					c.Message = err.Error()
					return
				}

				c.Health = health.OK
				c.Message = ""
			},
		},
	}
	registry.RegisterComponent(ctx, health.NewComponent("database", checks...))

	repository = &Repository{
		logger:   logger,
		registry: registry,
		config:   config,
		pool:     pool,
	}

	return repository, nil
}

func (r Repository) GetTodos(ctx context.Context) (t []todos.Todo, err error) {
	rows, err := r.pool.Query(ctx,
		`
		SELECT id, description, completed_at, created_at, updated_at
		FROM todos
		WHERE deleted_at IS NULL
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed querying database: %w", err)
	}

	// Use append to avoid returning nil slice
	t = make([]todos.Todo, 0)

	t, err = pgx.AppendRows(t, rows, pgx.RowToStructByName[todos.Todo])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDatabase, err)
	}

	return t, nil
}

func (r Repository) GetTodo(ctx context.Context, id uuid.UUID) (t todos.Todo, err error) {
	rows, err := r.pool.Query(ctx,
		`
		SELECT id, description, completed_at, created_at, updated_at
		FROM todos
		WHERE id=$1 AND deleted_at IS NULL
		`,
		id,
	)
	if err != nil {
		return todos.Todo{}, fmt.Errorf("failed querying database: %w", err)
	}

	t, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[todos.Todo])
	if errors.Is(err, pgx.ErrNoRows) {
		return todos.Todo{}, ErrTodoNotFound
	}

	if err != nil {
		return todos.Todo{}, fmt.Errorf("%w: %w", ErrDatabase, err)
	}

	return t, nil
}

func (r Repository) CreateTodo(ctx context.Context, todo todos.Todo) (createdTodo todos.Todo, err error) {
	rows, err := r.pool.Query(ctx,
		`
		INSERT INTO todos (description, created_at)
		VALUES ($1, $2)
		RETURNING id, description, completed_at, created_at, updated_at
		`,
		todo.Description,
		todo.CreatedAt,
	)
	if err != nil {
		return todos.Todo{}, fmt.Errorf("failed querying database: %w", err)
	}

	createdTodo, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[todos.Todo])
	if err != nil {
		return todos.Todo{}, fmt.Errorf("%w: %w", ErrDatabase, err)
	}

	return createdTodo, nil
}

func (r Repository) SaveTodo(ctx context.Context, todo todos.Todo) (savedTodo todos.Todo, err error) {
	rows, err := r.pool.Query(ctx,
		`
		UPDATE todos
		SET description = $2, completed_at = $3, updated_at = $4
		WHERE id=$1
		RETURNING id, description, completed_at, created_at, updated_at
		`,
		todo.ID,
		todo.Description,
		todo.CompletedAt,
		todo.UpdatedAt,
	)
	if err != nil {
		return todos.Todo{}, fmt.Errorf("failed querying database: %w", err)
	}

	savedTodo, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[todos.Todo])
	if errors.Is(err, pgx.ErrNoRows) {
		return todos.Todo{}, ErrTodoNotFound
	}

	if err != nil {
		return todos.Todo{}, fmt.Errorf("%w: %w", ErrDatabase, err)
	}

	return savedTodo, nil
}

func (r Repository) DeleteTodo(ctx context.Context, id uuid.UUID, deletedAt time.Time) error {
	c, err := r.pool.Exec(ctx,
		`
		UPDATE todos
		SET deleted_at=$2
		WHERE id=$1
		`,
		id,
		deletedAt,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDatabase, err)
	}

	if c.RowsAffected() == 0 {
		return ErrTodoNotFound
	}

	return nil
}
