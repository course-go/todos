package repository

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/todos"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // Used to register "pgx5" driver used for migrations.
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	}
	return
}

func (r *Repository) GetTodos(ctx context.Context) (t []todos.Todo, err error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM todos WHERE deleted_at IS NOT NULL")
	if err != nil {
		err = fmt.Errorf("failed querying database: %w", err)
		return
	}

	t, err = pgx.CollectRows(rows, pgx.RowTo[todos.Todo])
	r.logger.Info("got todos", "todos", t)
	return nil, nil
}

func (r *Repository) GetTodo(id uuid.UUID) (todo todos.Todo, err error) { //nolint
	return todos.Todo{}, nil
}

func (r *Repository) CreateTodo(todo todos.Todo) (createdTodo todos.Todo) { //nolint
	return todos.Todo{}
}

func (r *Repository) SaveTodo(todo todos.Todo) (savedTodo todos.Todo) { //nolint
	return todos.Todo{}
}

func (r *Repository) DeleteTodo(id uuid.UUID) (err error) { //nolint
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
			logger.Warn("failed closing migrations source: %w",
				"error", srcErr,
			)
		}

		if dbErr != nil {
			logger.Warn("failed closing database after migrations",
				"error", dbErr,
			)
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
