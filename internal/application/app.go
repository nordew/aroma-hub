package application

import (
	"aroma-hub/internal/application/service"
	"aroma-hub/internal/config"
	v1 "aroma-hub/internal/controller/http/v1"
	"aroma-hub/internal/infrastructure/adapters/storage"
	"aroma-hub/pkg/client/db/pgsql"
	"context"

	_ "aroma-hub/docs/api"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
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

	setSwagger(router)

	handler.MustInitAndRun(router, cfg.Server)
}

func setSwagger(router *fiber.App) {
	router.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
		Title:       "Aroma-Hub API",
	}))
}
