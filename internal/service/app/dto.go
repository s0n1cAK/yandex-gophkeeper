package app

import "yandex-gophkeeper/internal/domain"

type SecretUpsert struct {
	Type    domain.SecretType `json:"type"`
	Comment string            `json:"comment,omitempty"`
	Data    string            `json:"data"`
}
