package application

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "aroma-hub/docs/api"
	"aroma-hub/internal/application/service"
	"aroma-hub/internal/config"
	v1 "aroma-hub/internal/controller/http/v1"
	"aroma-hub/internal/infrastructure/adapters/messaging/telegram"
	"aroma-hub/internal/infrastructure/adapters/storage"
	"aroma-hub/internal/infrastructure/workers"
	"aroma-hub/pkg/auth"
	"aroma-hub/pkg/client/db/pgsql"
	"aroma-hub/pkg/otp_generator"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	pgxtransactor "github.com/nordew/pgx-transactor"

	stash "github.com/nordew/go-stash"
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

	transactor := pgxtransactor.NewTransactor(pool)

	otpGen := otp_generator.NewDefaultGenerator()
	cache := stash.NewCache()

	tokenCfg := auth.Config{
		AccessTokenSecret:    cfg.Auth.AuthSecret,
		AccessTokenDuration:  cfg.Auth.AccessTokenTTL,
		RefreshTokenSecret:   cfg.Auth.AuthSecret,
		RefreshTokenDuration: cfg.Auth.RefreshTokenTTL,
	}
	tokenService := auth.NewTokenService(tokenCfg)

	storages := storage.NewStorage(pool)

	telegramProvider, err := telegram.NewTelegramProvider(
		cfg.Telegram.Token,
		storages,
		otpGen,
		cache,
	)
	if err != nil {
		logger.Fatalf("Failed to create Telegram provider: %v", err)
	}
	services := service.NewService(storages, transactor, cache, tokenService, telegramProvider)

	promocodeWorker := workers.NewPromocodeWorker(services, logger)
	promocodeWorker.Start()

	slogHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slogLogger := slog.New(slogHandler)

	handler := v1.NewHandler(services, slogLogger, tokenService)
	router := createRouter(&cfg)
	setSwagger(router)

	go func() {
		cacheCfg := stash.CacheWorkerConfig{
			Cache:    cache,
			Interval: 1 * time.Minute,
			StopCh:   make(chan struct{}),
		}

		stash.StartCacheWorker(ctx, cacheCfg)
	}()

	telegramProvider.Start()

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
	telegramProvider.Stop()

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
