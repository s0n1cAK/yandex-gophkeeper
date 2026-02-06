package client

import (
	"time"
	customtype "yandex-gophkeeper/internal/customType/duration"
)

type Config struct {
	Address      string              `env:"ADDRESS" json:"address"`
	Timeout      customtype.Duration `env:"TIMEOUT" json:"timeout"`
	TokenFile    string              `env:"TOKEN_FILE" json:"token_file"`
	CryptoKey    string              `env:"CRYPTO_KEY" json:"crypto_key"`
	RetryMax     int                 `env:"RETRY_MAX" json:"retry_max"`
	RetryWaitMin customtype.Duration `env:"RETRY_WAIT_MIN" json:"retry_wait_min"`
	RetryWaitMax customtype.Duration `env:"RETRY_WAIT_MAX" json:"retry_wait_max"`
}

func Default() Config {
	return Config{
		Address:      "http://localhost:8080",
		Timeout:      customtype.Duration(5 * time.Second),
		TokenFile:    "gophkeeper.token",
		CryptoKey:    "",
		RetryMax:     3,
		RetryWaitMin: customtype.Duration(500 * time.Millisecond),
		RetryWaitMax: customtype.Duration(3 * time.Second),
	}
}
