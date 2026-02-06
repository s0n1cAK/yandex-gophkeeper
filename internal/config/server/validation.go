package server

import (
	"errors"
	"net"
	"strings"
)

var (
	ErrEmptyAddress     = errors.New("address is empty")
	ErrBadAddress       = errors.New("address is invalid")
	ErrEmptyDatabaseDSN = errors.New("database dsn is empty")
	ErrEmptyJWTSecret   = errors.New("jwt secret is empty")
	ErrBadTokenTTL      = errors.New("token ttl must be > 0")
)

func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.Address) == "" {
		return ErrEmptyAddress
	}
	if _, err := net.ResolveTCPAddr("tcp", cfg.Address); err != nil {
		return ErrBadAddress
	}

	if strings.TrimSpace(cfg.DatabaseDSN) == "" {
		return ErrEmptyDatabaseDSN
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return ErrEmptyJWTSecret
	}
	if cfg.TokenTTL.Duration() <= 0 {
		return ErrBadTokenTTL
	}

	return nil
}
