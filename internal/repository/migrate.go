package repository

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/course-go/todos/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Migrate(config *config.Database, logger *slog.Logger) error {
	databaseURL := fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		"pgx5", // golang-migrate uses "stdlib registered" drivers set by imports
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
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
