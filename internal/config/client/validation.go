package client

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrEmptyAddress   = errors.New("address is empty")
	ErrBadAddress     = errors.New("address is invalid (expected http(s)://host:port)")
	ErrBadTimeout     = errors.New("timeout must be > 0")
	ErrBadRetryMax    = errors.New("retry_max must be >= 0")
	ErrBadRetryWait   = errors.New("retry wait must be >= 0")
	ErrBadCryptoKey   = errors.New("cryptokey is empty")
	ErrBadInsecureTLS = errors.New("allowed only for https://localhost")
)

func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.Address) == "" {
		return ErrEmptyAddress
	}

	u, err := url.Parse(cfg.Address)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ErrBadAddress
	}

	if cfg.Timeout.Duration() <= 0 {
		return ErrBadTimeout
	}

	if cfg.RetryMax < 0 {
		return ErrBadRetryMax
	}

	if cfg.RetryWaitMin.Duration() < 0 || cfg.RetryWaitMax.Duration() < 0 {
		return ErrBadRetryWait
	}

	if cfg.InsecureSkipVerify {
		u, _ := url.Parse(cfg.Address)
		if u == nil || u.Scheme != "https" {
			return ErrBadInsecureTLS
		}
		h := strings.ToLower(u.Hostname())
		if h != "localhost" && h != "127.0.0.1" && h != "::1" {
			return ErrBadInsecureTLS
		}
	}
	return nil
}

func normalizeAddress(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	return "http://" + s
}
