package database

import (
	"database/sql"
	"fmt"
	"io/fs"

	"sprout-backend/internal/infrastructure/logger"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Migrator struct {
	db           *sql.DB
	migrationsFS fs.FS
}

func NewMigrator(db *sql.DB, migrationsFS fs.FS) *Migrator {
	return &Migrator{db: db, migrationsFS: migrationsFS}
}

func (m *Migrator) RunMigrations() error {
	logger.Info("Running database migrations...")

	goose.SetBaseFS(m.migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	currentVersion, err := goose.GetDBVersion(m.db)
	if err != nil && err != goose.ErrVersionNotFound {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	logger.Infof("Current database migration version: %d", currentVersion)

	if err := goose.Up(m.db, "."); err != nil {
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

func (m *Migrator) GetStatus() error {
	goose.SetBaseFS(m.migrationsFS)
	logger.Info("Migration status:")
	if err := goose.Status(m.db, "."); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}
	return nil
}
