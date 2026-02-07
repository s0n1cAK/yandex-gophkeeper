package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	config "yandex-gophkeeper/internal/config/server"
	"yandex-gophkeeper/internal/logger"
	app "yandex-gophkeeper/internal/service/app"
	postgres "yandex-gophkeeper/internal/storage/postgresql"

	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, err := logger.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init logger:%s", err)
		os.Exit(1)
	}
	defer func() {
		_ = logger.Sync()
	}()

	cfg, err := config.New()
	if errors.Is(err, config.ErrHelpRequested) {
		return
	}
	if err != nil {
		logger.Fatal("config error", zap.Error(err))
	}

	store, err := postgres.Init(ctx, cfg.DatabaseDSN, "migrations")
	if err != nil {
		logger.Fatal("postgres init failed", zap.Error(err))
	}
	defer store.Close()

	a, err := app.New(ctx, app.Deps{
		Config: &cfg,
		Logger: logger,
		Store:  store,
	})
	if err != nil {
		logger.Fatal("failed to init app", zap.Error(err))
	}
	defer a.Close()

	if err := a.Start(ctx); err != nil {
		logger.Fatal("app stopped with error", zap.Error(err))
	}
}
