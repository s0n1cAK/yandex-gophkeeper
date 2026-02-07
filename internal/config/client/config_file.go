package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type fileConfig struct {
	Address            *string `json:"address"`
	Timeout            *string `json:"timeout"`
	TokenFile          *string `json:"token_file"`
	RetryMax           *int    `json:"retry_max"`
	RetryWaitMin       *string `json:"retry_wait_min"`
	RetryWaitMax       *string `json:"retry_wait_max"`
	InsecureSkipVerify *bool   `json:"insecure_skip_verify"`
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
		cfg.Address = normalizeAddress(*fc.Address)
	}
	if fc.Timeout != nil {
		if err := cfg.Timeout.Set(*fc.Timeout); err != nil {
			return fmt.Errorf("bad timeout in config: %w", err)
		}
	}
	if fc.TokenFile != nil {
		cfg.TokenFile = *fc.TokenFile
	}
	if fc.RetryMax != nil {
		cfg.RetryMax = *fc.RetryMax
	}
	if fc.RetryWaitMin != nil {
		if err := cfg.RetryWaitMin.Set(*fc.RetryWaitMin); err != nil {
			return fmt.Errorf("bad retry_wait_min in config: %w", err)
		}
	}
	if fc.RetryWaitMax != nil {
		if err := cfg.RetryWaitMax.Set(*fc.RetryWaitMax); err != nil {
			return fmt.Errorf("bad retry_wait_max in config: %w", err)
		}
	}
	if fc.InsecureSkipVerify != nil {
		cfg.InsecureSkipVerify = *fc.InsecureSkipVerify
	}

	return nil
}
