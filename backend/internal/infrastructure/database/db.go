package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"sprout-backend/internal/config"
	"sprout-backend/internal/infrastructure/logger"
)

type Database struct {
	Pool *pgxpool.Pool
	DB   *sql.DB
}

func New(cfg *config.Config) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 2 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	dbConfig, err := pgx.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config for sql.DB: %w", err)
	}

	sqlDB := stdlib.OpenDB(*dbConfig)
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database with sql.DB: %w", err)
	}

	logger.Info("Database connection established")

	return &Database{
		Pool: pool,
		DB:   sqlDB,
	}, nil
}

func (db *Database) Close() {
	if db.DB != nil {
		db.DB.Close()
	}
	if db.Pool != nil {
		db.Pool.Close()
		logger.Info("Database connection closed")
	}
}

func (db *Database) GetPool() *pgxpool.Pool {
	return db.Pool
}

func (db *Database) GetDB() *sql.DB {
	return db.DB
}

func (db *Database) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
