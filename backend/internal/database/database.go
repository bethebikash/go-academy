package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database/sqlc"
)

// DB wraps pgxpool.Pool and the generated SQLC queries.
type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

// NewPool initializes a PostgreSQL connection pool using pgxpool.
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database connection URL: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DBMaxOpenConns)
	poolConfig.MinConns = int32(cfg.DBMaxIdleConns)
	poolConfig.MaxConnIdleTime = cfg.DBMaxIdleTime

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctxTimeout, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctxTimeout); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// NewDB creates a DB wrapper with initialized SQLC queries.
func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}

// Ping verifies database connectivity (satisfies health.DBPinger).
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Close gracefully releases the database connection pool.
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// RunMigrations executes pending golang-migrate database migrations.
func RunMigrations(databaseURL, migrationsPath string, logger *slog.Logger) error {
	// The pgx/v5 driver requires pgx5:// scheme for golang-migrate
	migrateURL := databaseURL
	if len(databaseURL) >= 11 && databaseURL[:11] == "postgres://" {
		migrateURL = "pgx5://" + databaseURL[11:]
	} else if len(databaseURL) >= 13 && databaseURL[:13] == "postgresql://" {
		migrateURL = "pgx5://" + databaseURL[13:]
	}

	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	m, err := migrate.New(sourceURL, migrateURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migration driver: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	if logger != nil {
		logger.Info("database migrations verified and up to date")
	}

	return nil
}
