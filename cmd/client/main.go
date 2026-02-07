package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"yandex-gophkeeper/internal/logger"
	"yandex-gophkeeper/internal/service/client"

	config "yandex-gophkeeper/internal/config/client"
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

	c, rest, err := config.LoadWithArgs(os.Args[1:])
	if errors.Is(err, config.ErrHelpRequested) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	var tok string
	if t, err := client.Read(c.TokenFile); err == nil {
		tok = t
	}

	cli := client.NewCLI(client.CLI{
		Client:    client.New(c, tok),
		TokenPath: c.TokenFile,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
	})

	os.Exit(cli.Run(ctx, rest))
}
