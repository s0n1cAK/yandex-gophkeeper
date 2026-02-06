package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"fmt"
	"io"
)

const NonceSize = 12

type Manager struct {
	aead cipher.AEAD
}

func NewManager(key []byte) (*Manager, error) {
	if key == nil {
		return nil, ErrNilKey
	}
	if len(key) != KeySize {
		return nil, ErrBadKeyLen
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypt: aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypt: cipher.NewGCM: %w", err)
	}
	if gcm.NonceSize() != NonceSize {
		return nil, fmt.Errorf("crypt: unexpected nonce size: got %d want %d", gcm.NonceSize(), NonceSize)
	}

	return &Manager{aead: gcm}, nil
}

func (m *Manager) Encrypt(plain, aad []byte) (nonce, ciphertext []byte, err error) {
	nonce = make([]byte, NonceSize)
	if _, err := io.ReadFull(crand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("crypt: read nonce: %w", err)
	}

	ciphertext = m.aead.Seal(nil, nonce, plain, aad)
	return nonce, ciphertext, nil
}

func (m *Manager) Decrypt(nonce, ciphertext, aad []byte) ([]byte, error) {
	if len(nonce) != NonceSize {
		return nil, ErrBadNonceLen
	}

	plain, err := m.aead.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, fmt.Errorf("crypt: gcm open: %w", err)
	}
	return plain, nil
}
