package server

import (
	"testing"
	"time"

	customtype "yandex-gophkeeper/internal/customType/duration"
)

func TestValidate_TLSCertBond(t *testing.T) {
	base := Config{
		Address:        ":8080",
		DatabaseDSN:    "postgres://u:p@localhost:5432/db?sslmode=disable",
		MigrationsPath: "migrations",
		JWTSecret:      "secret",
		TokenTTL:       customtype.Duration(24 * time.Hour),
		CryptoKey:      "keyfile",
		CryptoKeyID:    "v1",
	}

	onlyCert := base
	onlyCert.TLSCertPath = "cert.pem"
	onlyCert.TLSKeyPath = ""
	if err := Validate(onlyCert); err == nil {
		t.Fatalf("expected ErrBadCertBond")
	}

	both := base
	both.TLSCertPath = "cert.pem"
	both.TLSKeyPath = "key.pem"
	if err := Validate(both); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}
