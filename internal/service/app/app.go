package app

import (
	"context"
	"fmt"
	"net/http"

	config "yandex-gophkeeper/internal/config/server"
	"yandex-gophkeeper/internal/crypt"
	"yandex-gophkeeper/internal/service/secrets"
	postgres "yandex-gophkeeper/internal/storage/postgresql"
	transport "yandex-gophkeeper/internal/transport/http"
	"yandex-gophkeeper/internal/transport/http/handler"
	"yandex-gophkeeper/internal/transport/http/middleware"

	"go.uber.org/zap"
)

type Deps struct {
	Config *config.Config
	Logger *zap.Logger
	Store  *postgres.Store
}

type App struct {
	log   *zap.Logger
	store *postgres.Store
	http  *transport.Server
}

func New(ctx context.Context, d Deps) (*App, error) {
	if d.Config == nil {
		return nil, fmt.Errorf("app deps: config is nil")
	}
	if d.Logger == nil {
		return nil, fmt.Errorf("app deps: logger is nil")
	}
	if d.Store == nil {
		return nil, fmt.Errorf("app deps: store is nil")
	}

	pingH := handler.NewPing(d.Store)

	tokenMgr, err := NewTokenManager(d.Config.JWTSecret, d.Config.TokenTTL.Duration())
	if err != nil {
		return nil, fmt.Errorf("failed to init jwtmanager")
	}
	authSvc := NewAuthService(d.Store, tokenMgr)
	authH := handler.NewAuth(d.Logger, authSvc)

	authMW := middleware.Auth(tokenMgr)

	key, err := crypt.LoadAESKeyFromFile(d.Config.CryptoKey)
	if err != nil {
		return nil, fmt.Errorf("load aes key: %w", err)
	}
	kr, err := crypt.NewKeyring(d.Config.CryptoKeyID, key)
	if err != nil {
		return nil, fmt.Errorf("init keyring: %w", err)
	}

	secretsSvc := secrets.New(d.Store, kr)

	secretsH := handler.NewSecrets(d.Logger, secretsSvc)

	httpSrv, err := transport.New(transport.Deps{
		Config:  d.Config,
		Logger:  d.Logger,
		Auth:    authH,
		Secrets: secretsH,
		Ping:    pingH,
		AuthMW:  authMW,
	})
	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	return &App{
		log:   d.Logger,
		store: d.Store,
		http:  httpSrv,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.http.Start()
	}()

	select {
	case <-ctx.Done():
		if err := a.http.Shutdown(context.Background()); err != nil {
			a.log.Warn("http shutdown failed", zap.Error(err))
		}

		err := <-errCh
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil

	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

func (a *App) Close() {
	if a.store != nil {
		a.store.Close()
	}
}
