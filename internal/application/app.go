package application

import (
	_ "aroma-hub/docs/api"
	"aroma-hub/internal/application/service"
	"aroma-hub/internal/config"
	v1 "aroma-hub/internal/controller/http/v1"
	"aroma-hub/internal/infrastructure/adapters/storage"
	"aroma-hub/internal/infrastructure/workers"
	"aroma-hub/pkg/client/db/pgsql"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func MustRun() {
	logger := log.New(os.Stdout, "[APP] ", log.LstdFlags)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	cfg := config.MustLoad()
	pool := pgsql.MustConnect(ctx, cfg.Postgres)

	if cfg.Postgres.Migrate {
		pgsql.MustMigrate(ctx, pool, cfg.Postgres)
	}

	storages := storage.NewStorage(pool)
	services := service.NewService(storages)

	promocodeWorker := workers.NewPromocodeWorker(services, logger)
	promocodeWorker.Start()

	handler := v1.NewHandler(services)
	router := createRouter(&cfg)
	setSwagger(router)

	serverShutdownDone := make(chan struct{})
	go func() {
		defer close(serverShutdownDone)

		if err := handler.InitAndServe(router, cfg.Server); err != nil {
			logger.Printf("Server error: %v", err)
		}
	}()

	<-signalChan
	logger.Println("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	logger.Println("Gracefully shutting down...")

	logger.Println("Stopping worker...")
	promocodeWorker.Stop()

	logger.Println("Stopping HTTP server...")
	if err := router.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Printf("HTTP server shutdown error: %v", err)
	}

	<-serverShutdownDone

	logger.Println("Closing database connections...")
	pool.Close()

	logger.Println("Shutdown complete")
}

func createRouter(cfg *config.Config) *fiber.App {
	return fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})
}

func setSwagger(router *fiber.App) {
	router.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
		Title:       "Aroma-Hub API",
	}))
}
