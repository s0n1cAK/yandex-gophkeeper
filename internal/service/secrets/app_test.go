package secrets

import (
	"bytes"
	"context"
	"testing"

	"yandex-gophkeeper/internal/domain"
)

type fakeStore struct {
	addCalled bool
	addArgs   struct {
		uid     domain.UserID
		t       domain.SecretType
		comment string
		keyID   string
		nonce   []byte
		ct      []byte
	}

	getSecret domain.Secret
	getErr    error
}

func (s *fakeStore) AddSecret(ctx context.Context, userID domain.UserID, secretType domain.SecretType, comment, keyID string, nonce, ciphertext []byte) (domain.SecretID, error) {
	s.addCalled = true
	s.addArgs.uid = userID
	s.addArgs.t = secretType
	s.addArgs.comment = comment
	s.addArgs.keyID = keyID
	s.addArgs.nonce = append([]byte(nil), nonce...)
	s.addArgs.ct = append([]byte(nil), ciphertext...)
	return 123, nil
}

func (s *fakeStore) ListSecrets(ctx context.Context, userID domain.UserID) ([]domain.SecretMeta, error) {
	panic("not used")
}
func (s *fakeStore) GetSecret(ctx context.Context, userID domain.UserID, secretID domain.SecretID) (domain.Secret, error) {
	return s.getSecret, s.getErr
}
func (s *fakeStore) UpdateSecret(ctx context.Context, userID domain.UserID, secretID domain.SecretID, secretType domain.SecretType, comment, keyID string, nonce, ciphertext []byte) error {
	panic("not used")
}
func (s *fakeStore) DeleteSecret(ctx context.Context, userID domain.UserID, secretID domain.SecretID) error {
	panic("not used")
}

type fakeCrypto struct {
	encAAD []byte
	decAAD []byte

	nonce []byte
	ct    []byte
	plain []byte
}

func (c *fakeCrypto) Encrypt(plain, aad []byte) (keyID string, nonce, ciphertext []byte, err error) {
	c.encAAD = append([]byte(nil), aad...)
	c.plain = append([]byte(nil), plain...)
	return "v1", append([]byte(nil), c.nonce...), append([]byte(nil), c.ct...), nil
}

func (c *fakeCrypto) Decrypt(keyID string, nonce, ciphertext, aad []byte) ([]byte, error) {
	c.decAAD = append([]byte(nil), aad...)
	return []byte("PLAINTEXT"), nil
}

func TestService_Create_UsesAADAndStoresEncrypted(t *testing.T) {
	st := &fakeStore{}
	cr := &fakeCrypto{
		nonce: bytes.Repeat([]byte{1}, 12),
		ct:    []byte("CIPHERTEXT"),
	}

	svc := New(st, cr)

	uid := domain.UserID(7)
	in := SecretUpsert{
		Type:    domain.SecretText,
		Comment: "note",
		Data:    []byte("hello"),
	}

	id, err := svc.Create(context.Background(), uid, in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 123 {
		t.Fatalf("id=%d want 123", id)
	}
	if !st.addCalled {
		t.Fatalf("expected AddSecret call")
	}

	wantAAD := aad(uid, in.Type, in.Comment)
	if !bytes.Equal(cr.encAAD, wantAAD) {
		t.Fatalf("aad mismatch")
	}

	if st.addArgs.keyID != "v1" {
		t.Fatalf("keyID=%q", st.addArgs.keyID)
	}
	if !bytes.Equal(st.addArgs.nonce, cr.nonce) {
		t.Fatalf("nonce mismatch")
	}
	if !bytes.Equal(st.addArgs.ct, cr.ct) {
		t.Fatalf("ciphertext mismatch")
	}
}

func TestService_Get_UsesAAD(t *testing.T) {
	st := &fakeStore{
		getSecret: domain.Secret{
			ID:         1,
			OwnerID:    7,
			Type:       domain.SecretText,
			Comment:    "c",
			KeyID:      "v1",
			Nonce:      bytes.Repeat([]byte{9}, 12),
			Ciphertext: []byte("x"),
		},
	}
	cr := &fakeCrypto{}

	svc := New(st, cr)

	uid := domain.UserID(7)
	sec, plain, err := svc.Get(context.Background(), uid, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sec.ID != 1 {
		t.Fatalf("secret mismatch")
	}
	if string(plain) != "PLAINTEXT" {
		t.Fatalf("plain mismatch: %q", plain)
	}

	wantAAD := aad(uid, sec.Type, sec.Comment)
	if !bytes.Equal(cr.decAAD, wantAAD) {
		t.Fatalf("aad mismatch")
	}
}
