package server

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type fileConfig struct {
	Address        *string `json:"address"`
	DatabaseDSN    *string `json:"database_dsn"`
	MigrationsPath *string `json:"migrations_path"`
	JWTSecret      *string `json:"jwt_secret"`
	TokenTTL       *string `json:"token_ttl"`
	CryptoKey      *string `json:"crypto_key"`
	CryptoKeyID    *string `json:"crypto_key_id"`
	TLSCertPath    *string `json:"tls_cert"`
	TLSKeyPath     *string `json:"tls_key"`
}

func resolveConfigPath(args []string) (string, error) {
	for i := 0; i < len(args); i++ {
		a := args[i]

		switch a {
		case "-c", "-config", "--config":
			if i+1 < len(args) {
				return args[i+1], nil
			}
			return "", fmt.Errorf("%s requires a value", a)
		}

		if strings.HasPrefix(a, "-c=") {
			return strings.TrimPrefix(a, "-c="), nil
		}
		if strings.HasPrefix(a, "-config=") {
			return strings.TrimPrefix(a, "-config="), nil
		}
		if strings.HasPrefix(a, "--config=") {
			return strings.TrimPrefix(a, "--config="), nil
		}
	}

	if envPath := os.Getenv("CONFIG"); envPath != "" {
		return envPath, nil
	}
	return "", nil
}

func loadFile(path string) (fileConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, fmt.Errorf("read config file: %w", err)
	}
	var fc fileConfig
	if err := json.Unmarshal(b, &fc); err != nil {
		return fileConfig{}, fmt.Errorf("unmarshal json: %w", err)
	}
	return fc, nil
}

func applyFile(cfg *Config, fc fileConfig) error {
	if fc.Address != nil {
		cfg.Address = *fc.Address
	}
	if fc.DatabaseDSN != nil {
		cfg.DatabaseDSN = *fc.DatabaseDSN
	}
	if fc.MigrationsPath != nil {
		cfg.MigrationsPath = *fc.MigrationsPath
	}
	if fc.JWTSecret != nil {
		cfg.JWTSecret = *fc.JWTSecret
	}
	if fc.TokenTTL != nil {
		if err := cfg.TokenTTL.Set(*fc.TokenTTL); err != nil {
			return fmt.Errorf("bad token_ttl in config: %w", err)
		}
	}
	if fc.CryptoKey != nil {
		cfg.CryptoKey = *fc.CryptoKey
	}
	if fc.CryptoKeyID != nil {
		cfg.CryptoKeyID = *fc.CryptoKeyID
	}
	if fc.TLSCertPath != nil {
		cfg.TLSCertPath = *fc.TLSCertPath
	}
	if fc.TLSKeyPath != nil {
		cfg.TLSKeyPath = *fc.TLSKeyPath
	}
	return nil
}
