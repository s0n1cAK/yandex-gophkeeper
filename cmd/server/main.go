package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	config "yandex-gophkeeper/internal/config/server"
	"yandex-gophkeeper/internal/logger"
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
		log.Fatal("config error", zap.Error(err))
	}

	store, err := postgres.Init(ctx, cfg.DatabaseDSN, "migrations")
	if err != nil {
		logger.Fatal("postgres init failed", zap.Error(err))
	}
	defer store.Close()

	fmt.Println(cfg)
}
