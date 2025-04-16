package application

import (
	"aroma-hub/internal/application/service"
	"aroma-hub/internal/config"
	v1 "aroma-hub/internal/controller/http/v1"
	"aroma-hub/internal/infrastructure/adapters/storage"
	pgsql "aroma-hub/pkg/client/db"
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

	productStorage := storage.NewProductStorage(pool)
	categoryStorage := storage.NewCategoryStorage(pool)

	productService := service.NewProductService(productStorage, categoryStorage)

	handler := v1.NewHandler(productService)

	router := fiber.New()

	handler.MustInitAndRun(router, cfg.Server)
}
