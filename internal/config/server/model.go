package server

import (
	"time"
	customtype "yandex-gophkeeper/internal/customType/duration"
)

type Config struct {
	Address        string              `env:"RUN_ADDRESS" json:"address"`
	DatabaseDSN    string              `env:"DATABASE_DSN" json:"database_dsn"`
	MigrationsPath string              `env:"MIGRATIONS_PATH" json:"migrations_path"`
	JWTSecret      string              `env:"JWT_SECRET" json:"jwt_secret"`
	TokenTTL       customtype.Duration `env:"TOKEN_TTL" json:"token_ttl"`
	CryptoKey      string              `env:"CRYPTO_KEY" json:"crypto_key"`
	CryptoKeyID    string              `env:"CRYPTO_KEY_ID" json:"crypto_key_id"`
	TLSCertPath    string              `env:"TLS_CERT" json:"tls_cert"`
	TLSKeyPath     string              `env:"TLS_KEY" json:"tls_key"`
}

func Default() Config {
	return Config{
		Address:        ":8080",
		DatabaseDSN:    "",
		MigrationsPath: "migrations",
		JWTSecret:      "",
		TokenTTL:       customtype.Duration(24 * time.Hour),
		CryptoKey:      "",
		CryptoKeyID:    "v1",
		TLSCertPath:    "",
		TLSKeyPath:     "",
	}
}
