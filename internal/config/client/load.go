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
	cfg, _, err := LoadWithArgs(args)
	return cfg, err
}

func LoadWithArgs(args []string) (Config, []string, error) {
	cfg := Default()

	cfgPath, err := resolveConfigPath(args)
	if err != nil {
		return Config{}, nil, err
	}
	if cfgPath != "" {
		fc, err := loadFile(cfgPath)
		if err != nil {
			return Config{}, nil, err
		}
		if err := applyFile(&cfg, fc); err != nil {
			return Config{}, nil, err
		}
	}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, nil, fmt.Errorf("parse env: %w", err)
	}

	fs := flag.NewFlagSet("gophkeeper-agent", flag.ContinueOnError)

	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", os.Args[0])
		fs.PrintDefaults()
	}

	fs.StringVar(&cfg.Address, "a", cfg.Address, "Server address")
	fs.BoolVar(&cfg.InsecureSkipVerify, "insecure-skip-verify", cfg.InsecureSkipVerify, "Skip TLS cert verification")
	fs.Var(&cfg.Timeout, "timeout", "Request timeout")
	fs.StringVar(&cfg.TokenFile, "token-file", cfg.TokenFile, "Path to local token file")

	fs.IntVar(&cfg.RetryMax, "retry-max", cfg.RetryMax, "Max retries for failed requests")
	fs.Var(&cfg.RetryWaitMin, "retry-wait-min", "Min retry wait")
	fs.Var(&cfg.RetryWaitMax, "retry-wait-max", "Max retry wait")

	fs.String("c", "", "Path to config file (JSON)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, nil, ErrHelpRequested
		}
		return Config{}, nil, err
	}

	cfg.Address = normalizeAddress(cfg.Address)

	if err := Validate(cfg); err != nil {
		return Config{}, nil, err
	}

	return cfg, fs.Args(), nil
}

func New() (Config, error) {
	return Load(os.Args[1:])
}
