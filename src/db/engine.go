package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Config encapsulates database connection pool parameters.
type Config struct {
	DriverName      string
	DataSourceName  string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DBTX abstracts execution between standard SQL connection pools and active transactions.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Engine manages the root SQL connection pool lifecycle.
type Engine struct {
	db *sql.DB
}

// NewEngine establishes and verifies the database connection pool using the provided configuration.
func NewEngine(ctx context.Context, cfg Config) (*Engine, error) {
	if cfg.DriverName == "" || cfg.DataSourceName == "" {
		return nil, fmt.Errorf("engine: driver name and datasource name must not be empty")
	}

	conn, err := sql.Open(cfg.DriverName, cfg.DataSourceName)
	if err != nil {
		return nil, fmt.Errorf("engine: failed to open database: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		conn.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		conn.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		conn.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := conn.PingContext(pingCtx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("engine: failed to verify database connectivity: %w", err)
	}

	return &Engine{db: conn}, nil
}

// Close gracefully terminates all idle and active connections in the pool.
func (e *Engine) Close() error {
	if e.db == nil {
		return nil
	}
	return e.db.Close()
}

// DB returns the underlying connection pool handle.
func (e *Engine) DB() *sql.DB {
	return e.db
}
