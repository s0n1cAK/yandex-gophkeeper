package client

import (
	"os"
	"strings"
)

func Read(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func Write(path, token string) error {
	token = strings.TrimSpace(token)
	return os.WriteFile(path, []byte(token+"\n"), 0o600)
}
