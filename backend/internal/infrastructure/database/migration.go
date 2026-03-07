package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"sprout-backend/internal/infrastructure/logger"
)

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) RunMigrations(migrationsDir string) error {
	logger.Info("Running database migrations...")

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	currentVersion, err := goose.GetDBVersion(m.db)
	if err != nil && err != goose.ErrVersionNotFound {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	logger.Infof("Current database migration version: %d", currentVersion)

	if err := goose.Up(m.db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	newVersion, err := goose.GetDBVersion(m.db)
	if err != nil {
		return fmt.Errorf("failed to get new migration version: %w", err)
	}

	if currentVersion != newVersion {
		logger.Infof("Database migrated to version: %d", newVersion)
	} else {
		logger.Info("Database is already up to date")
	}

	return nil
}

func (m *Migrator) GetStatus(migrationsDir string) error {
	logger.Info("Migration status:")
	if err := goose.Status(m.db, migrationsDir); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}
	return nil
}
