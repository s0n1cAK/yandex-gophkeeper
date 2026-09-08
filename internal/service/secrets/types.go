package secrets

import "yandex-gophkeeper/internal/domain"

type SecretUpsert struct {
	Type    domain.SecretType
	Comment string
	Data    []byte
}
