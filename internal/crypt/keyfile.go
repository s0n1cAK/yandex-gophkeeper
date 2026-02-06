package crypt

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

const KeySize = 32

func LoadAESKeyFromFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("crypt: read key file: %w", err)
	}

	s := strings.TrimSpace(string(b))
	raw := []byte(s)

	if len(raw) == KeySize {
		out := make([]byte, KeySize)
		copy(out, raw)
		return out, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("crypt: decode base64 key: %w", err)
	}
	if len(decoded) != KeySize {
		return nil, ErrBadKeyLen
	}

	out := make([]byte, KeySize)
	copy(out, decoded)
	return out, nil
}
