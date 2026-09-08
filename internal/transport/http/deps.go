package http

import (
	"errors"
	"net/http"

	srvconfig "yandex-gophkeeper/internal/config/server"

	"go.uber.org/zap"
)

var (
	ErrNilConfig  = errors.New("deps: config is nil")
	ErrNilLogger  = errors.New("deps: logger is nil")
	ErrNilAuth    = errors.New("deps: auth handler is nil")
	ErrNilSecrets = errors.New("deps: secrets handler is nil")
	ErrNilPing    = errors.New("deps: ping handler is nil")
	ErrNilAuthMW  = errors.New("deps: auth middleware is nil")
)

type Deps struct {
	Config *srvconfig.Config
	Logger *zap.Logger

	Auth    AuthHandler
	Secrets SecretsHandler
	Ping    PingHandler

	AuthMW func(http.Handler) http.Handler
}

func (d Deps) Validate() error {
	if d.Config == nil {
		return ErrNilConfig
	}
	if d.Logger == nil {
		return ErrNilLogger
	}
	if d.Auth == nil {
		return ErrNilAuth
	}
	if d.Secrets == nil {
		return ErrNilSecrets
	}
	if d.Ping == nil {
		return ErrNilPing
	}
	if d.AuthMW == nil {
		return ErrNilAuthMW
	}
	return nil
}

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

type SecretsHandler interface {
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type PingHandler interface {
	Ping(w http.ResponseWriter, r *http.Request)
}
