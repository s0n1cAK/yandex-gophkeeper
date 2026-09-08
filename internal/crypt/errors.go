package crypt

import "errors"

var (
	ErrNilKey      = errors.New("crypt: key is nil")
	ErrBadKeyLen   = errors.New("crypt: key must be 32 bytes (AES-256)")
	ErrBadNonceLen = errors.New("crypt: bad nonce length")
)
