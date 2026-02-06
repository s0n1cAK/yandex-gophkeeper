package domain

import "time"

type SecretID int64

type SecretType string

const (
	SecretPassword SecretType = "password"
	SecretText     SecretType = "text"
	SecretBankCard SecretType = "bank_card"
	SecretBinary   SecretType = "binary"
)

type Secret struct {
	ID         SecretID   `json:"id"`
	OwnerID    UserID     `json:"owner_id"`
	Type       SecretType `json:"type"`
	Comment    string     `json:"comment,omitempty"`
	KeyID      string     `json:"-"`
	Nonce      []byte     `json:"-"`
	Ciphertext []byte     `json:"-"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type SecretMeta struct {
	ID        SecretID   `json:"id"`
	Type      SecretType `json:"type"`
	Comment   string     `json:"comment,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func ValidateSecretType(t SecretType) error {
	switch t {
	case SecretPassword, SecretText, SecretBankCard, SecretBinary:
		return nil
	default:
		return ErrInvalidSecretType
	}
}
