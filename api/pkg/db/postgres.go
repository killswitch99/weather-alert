package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool *pgxpool.Pool

	ErrNoRows = sql.ErrNoRows

	defaultQueryTimeout = 10 * time.Second
)

type Config struct {
	URI             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	QueryTimeout    time.Duration
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
		QueryTimeout:    defaultQueryTimeout,
	}
}

func Connect(config *Config) error {
	var err error
	pool, err = pgxpool.New(context.Background(), config.URI)
	if err != nil {
		return fmt.Errorf("failed to create pgx pool: %w", err)
	}

	pool.Config().MaxConns = int32(config.MaxOpenConns)
	pool.Config().MaxConnIdleTime = config.ConnMaxLifetime

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return nil
}

// Disconnect gracefully closes the pool
func Disconnect() {
	if pool != nil {
		pool.Close()
	}
}

// GetPool gives callers access to the underlying pool
func GetPool() *pgxpool.Pool {
	return pool
}

func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, defaultQueryTimeout)
}

func HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return pool.Ping(ctx)
}

func WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
