package crypt

import "fmt"

var ErrUnknownKeyID = fmt.Errorf("crypt: unknown key id")

type Keyring struct {
	current string
	keys    map[string]*Manager
}

func NewKeyring(currentID string, currentKey []byte) (*Keyring, error) {
	if currentID == "" {
		return nil, fmt.Errorf("crypt: empty current key id")
	}
	cm, err := NewManager(currentKey)
	if err != nil {
		return nil, err
	}
	return &Keyring{
		current: currentID,
		keys: map[string]*Manager{
			currentID: cm,
		},
	}, nil
}

func (k *Keyring) Encrypt(plain, aad []byte) (keyID string, nonce, ciphertext []byte, err error) {
	m := k.keys[k.current]
	n, ct, err := m.Encrypt(plain, aad)
	if err != nil {
		return "", nil, nil, err
	}
	return k.current, n, ct, nil
}

func (k *Keyring) Decrypt(keyID string, nonce, ciphertext, aad []byte) ([]byte, error) {
	m := k.keys[keyID]
	if m == nil {
		return nil, fmt.Errorf("%w: %s", ErrUnknownKeyID, keyID)
	}
	return m.Decrypt(nonce, ciphertext, aad)
}
