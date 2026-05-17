package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/relationskat/auth-service/internal/config"
)

const createTableStmt = `
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY,
    last_name       VARCHAR(255) NOT NULL,
    first_name      VARCHAR(255) NOT NULL,
    middle_name     VARCHAR(255),
    email           VARCHAR(255) NOT NULL,
    email_confirmed BOOLEAN      NOT NULL DEFAULT FALSE,
    password_hash   VARCHAR(255) NOT NULL,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ
)`

const createIndexStmt = `
CREATE UNIQUE INDEX IF NOT EXISTS users_email_active_uniq ON users(email) WHERE deleted_at IS NULL`

func dsn(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DB,
	)
}

func NewPostgres(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool

	err := retry(10, time.Second, func() error {
		p, err := pgxpool.New(ctx, dsn(cfg))
		if err != nil {
			return err
		}
		if err := p.Ping(ctx); err != nil {
			p.Close()
			return err
		}
		pool = p
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, createTableStmt); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if _, err := pool.Exec(ctx, createIndexStmt); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
