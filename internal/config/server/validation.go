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
	ErrBadCryptoKey     = errors.New("cryptokey is empty")
	ErrBadCryptoKeyID   = errors.New("cryptokeyid is empty")
	ErrBadCertBond      = errors.New("tls_cert and tls_key must be set together")
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

	if strings.TrimSpace(cfg.CryptoKey) == "" {
		return ErrBadCryptoKey
	}
	if strings.TrimSpace(cfg.CryptoKeyID) == "" {
		return ErrBadCryptoKeyID
	}

	c := strings.TrimSpace(cfg.TLSCertPath)
	k := strings.TrimSpace(cfg.TLSKeyPath)
	if (c == "") != (k == "") {
		return ErrBadCertBond
	}
	return nil
}
