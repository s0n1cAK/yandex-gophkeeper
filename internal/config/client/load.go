package client

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

	fs := flag.NewFlagSet("gophkeeper-agent", flag.ContinueOnError)

	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", os.Args[0])
		fs.PrintDefaults()
	}

	fs.StringVar(&cfg.Address, "a", cfg.Address, "Server address, e.g. http://host:port")
	fs.Var(&cfg.Timeout, "timeout", "Request timeout, e.g. 5s (or seconds as number)")
	fs.StringVar(&cfg.TokenFile, "token-file", cfg.TokenFile, "Path to local token file")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to public key (PEM) (optional)")

	fs.IntVar(&cfg.RetryMax, "retry-max", cfg.RetryMax, "Max retries for failed requests")
	fs.Var(&cfg.RetryWaitMin, "retry-wait-min", "Min retry wait, e.g. 500ms")
	fs.Var(&cfg.RetryWaitMax, "retry-wait-max", "Max retry wait, e.g. 3s")

	fs.String("c", "", "Path to config file (JSON)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, ErrHelpRequested
		}
		return Config{}, err
	}

	cfg.Address = normalizeAddress(cfg.Address)

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func New() (Config, error) {
	return Load(os.Args[1:])
}
