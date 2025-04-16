package pgsql

import (
	"aroma-hub/internal/config"
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
	"log"
)

func MustConnect(ctx context.Context, cfg config.Postgres) *pgxpool.Pool {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		log.Fatalf("failed to parse dsn: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
	}

	return pool
}

func MustMigrate(
	ctx context.Context,
	pool *pgxpool.Pool,
	cfg config.Postgres,
) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	db := stdlib.OpenDBFromPool(pool)
	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			log.Fatalf("failed to close db connection: %v", err)
		}
	}(db)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	if err := goose.Up(db, cfg.MigrationsDir); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	log.Println("migrations applied successfully")
}
