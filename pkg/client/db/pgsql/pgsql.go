package pgsql

import (
	"aroma-hub/internal/config"
	"context"
	"database/sql"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

func MustConnect(ctx context.Context, cfg config.Postgres) *pgxpool.Pool {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		log.Fatalf("failed to parse dsn: %v", err)
	}

	poolCfg.ConnConfig.Tracer = &pgxTracer{}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Printf("Connected to PostgreSQL database: %s", hidePassword(cfg.DSN))
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

	log.Printf("Applying migrations from directory: %s", cfg.MigrationsDir)

	if err := goose.Up(db, cfg.MigrationsDir); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		log.Printf("Warning: couldn't get current migration version: %v", err)
	} else {
		log.Printf("Migrations applied successfully. Current version: %d", currentVersion)
	}
}

func hidePassword(dsn string) string {
	result := dsn

	for i := 0; i < len(dsn); i++ {
		if dsn[i] == ':' {
			for j := i + 1; j < len(dsn); j++ {
				if dsn[j] == '@' {
					result = dsn[:i+1] + "****" + dsn[j:]
					break
				}
			}
			break
		}
	}

	return result
}
