package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	config "yandex-gophkeeper/internal/config/server"

	"go.uber.org/zap"
)

type Server struct {
	cfg    *config.Config
	log    *zap.Logger
	router http.Handler
	srv    *http.Server
}

func New(d Deps) (*Server, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}

	r := buildRouter(d)

	httpSrv := &http.Server{
		Addr:              d.Config.Address,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Server{
		cfg:    d.Config,
		log:    d.Logger,
		router: r,
		srv:    httpSrv,
	}, nil
}

func (s *Server) Start() error {
	s.log.Info("starting server", zap.String("addr", s.cfg.Address))
	err := s.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http listen: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctxContextDone <-chan struct{}) {
	<-ctxContextDone
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.srv.Shutdown(ctx)
}
