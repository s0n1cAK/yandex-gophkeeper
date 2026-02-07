package secrets

import (
	"context"
	"crypto/sha256"
	"fmt"

	"yandex-gophkeeper/internal/domain"
)

type Store interface {
	AddSecret(ctx context.Context, userID domain.UserID, secretType domain.SecretType, comment, keyID string, nonce, ciphertext []byte) (domain.SecretID, error)
	ListSecrets(ctx context.Context, userID domain.UserID) ([]domain.SecretMeta, error)
	GetSecret(ctx context.Context, userID domain.UserID, secretID domain.SecretID) (domain.Secret, error)
	UpdateSecret(ctx context.Context, userID domain.UserID, secretID domain.SecretID, secretType domain.SecretType, comment, keyID string, nonce, ciphertext []byte) error
	DeleteSecret(ctx context.Context, userID domain.UserID, secretID domain.SecretID) error
}

type Crypto interface {
	Encrypt(plain, aad []byte) (keyID string, nonce, ciphertext []byte, err error)
	Decrypt(keyID string, nonce, ciphertext, aad []byte) ([]byte, error)
}

type Service struct {
	st    Store
	crypt Crypto
}

func New(st Store, crypt Crypto) *Service {
	return &Service{st: st, crypt: crypt}
}

func (s *Service) Create(ctx context.Context, ownerID domain.UserID, in SecretUpsert) (domain.SecretID, error) {
	aad := aad(ownerID, in.Type, in.Comment)
	keyID, nonce, ct, err := s.crypt.Encrypt(in.Data, aad)
	if err != nil {
		return 0, fmt.Errorf("encrypt: %w", err)
	}
	return s.st.AddSecret(ctx, ownerID, in.Type, in.Comment, keyID, nonce, ct)
}

func (s *Service) List(ctx context.Context, ownerID domain.UserID) ([]domain.SecretMeta, error) {
	return s.st.ListSecrets(ctx, ownerID)
}

func (s *Service) Get(ctx context.Context, ownerID domain.UserID, id domain.SecretID) (domain.Secret, []byte, error) {
	sec, err := s.st.GetSecret(ctx, ownerID, id)
	if err != nil {
		return domain.Secret{}, nil, err
	}
	plain, err := s.crypt.Decrypt(sec.KeyID, sec.Nonce, sec.Ciphertext, aad(ownerID, sec.Type, sec.Comment))
	if err != nil {
		return domain.Secret{}, nil, fmt.Errorf("decrypt: %w", err)
	}
	return sec, plain, nil
}

func (s *Service) Update(ctx context.Context, ownerID domain.UserID, id domain.SecretID, in SecretUpsert) error {
	keyID, nonce, ct, err := s.crypt.Encrypt(in.Data, aad(ownerID, in.Type, in.Comment))
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}
	return s.st.UpdateSecret(ctx, ownerID, id, in.Type, in.Comment, keyID, nonce, ct)
}

func (s *Service) Delete(ctx context.Context, ownerID domain.UserID, id domain.SecretID) error {
	return s.st.DeleteSecret(ctx, ownerID, id)
}

func aad(ownerID domain.UserID, t domain.SecretType, comment string) []byte {
	s := fmt.Sprintf("owner:%d|type:%s|comment:%s", ownerID, string(t), comment)
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}
