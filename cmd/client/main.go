package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	config "yandex-gophkeeper/internal/config/client"
	"yandex-gophkeeper/internal/logger"

	"go.uber.org/zap"
)

func main() {
	_, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
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

	fmt.Println(cfg)
}
