package server

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

var ErrHelpRequested = errors.New("help requested")

func Load(args []string) (Config, error) {
	cfg := Default()

	cfgPath, err := resolveConfigPath(args)
	if err != nil {
		return Config{}, err
	}
	if cfgPath != "" {
		fc, err := loadFile(cfgPath)
		if err != nil {
			return Config{}, err
		}
		if err := applyFile(&cfg, fc); err != nil {
			return Config{}, err
		}
	}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse env: %w", err)
	}

	fs := flag.NewFlagSet("gophkeeper-server", flag.ContinueOnError)

	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", os.Args[0])
		fs.PrintDefaults()
	}

	fs.StringVar(&cfg.Address, "a", cfg.Address, "HTTP listen address")
	fs.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "PostgreSQL DSN")
	fs.StringVar(&cfg.MigrationsPath, "migrations", cfg.MigrationsPath, "Migrations path")
	fs.StringVar(&cfg.JWTSecret, "jwt-secret", cfg.JWTSecret, "JWT secret")
	fs.Var(&cfg.TokenTTL, "token-ttl", "Token TTL")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to private key (AES)")

	fs.String("c", "", "Path to config file (JSON)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, ErrHelpRequested
		}
		return Config{}, err
	}

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func New() (Config, error) {
	return Load(os.Args[1:])
}
