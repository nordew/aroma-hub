package application

import (
	"aroma-hub/internal/application/service"
	"aroma-hub/internal/config"
	v1 "aroma-hub/internal/controller/http/v1"
	"aroma-hub/internal/infrastructure/adapters/storage"
	"aroma-hub/pkg/client/db/pgsql"
	"context"
	"github.com/gofiber/fiber/v2"
)

func MustRun() {
	ctx := context.Background()

	cfg := config.MustLoad()

	pool := pgsql.MustConnect(ctx, cfg.Postgres)

	if cfg.Postgres.Migrate {
		pgsql.MustMigrate(ctx, pool, cfg.Postgres)
	}

	storages := storage.NewStorage(pool)

	services := service.NewService(storages)

	handler := v1.NewHandler(services)

	router := fiber.New()

	handler.MustInitAndRun(router, cfg.Server)
}
