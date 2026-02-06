package postgres

import (
	"context"
	"fmt"
	"time"
	"yandex-gophkeeper/internal/domain"
)

func (s *Store) AddSecret(
	ctx context.Context,
	userID domain.UserID,
	secretType domain.SecretType,
	comment string,
	keyID string,
	nonce []byte,
	ciphertext []byte,
) (domain.SecretID, error) {
	const q = `
		INSERT INTO secrets(owner_id, type, comment, key_id, nonce, ciphertext, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id;
	`

	now := time.Now().UTC()

	var id int64
	err := withRetryWrite(ctx, func() error {
		return s.pool.QueryRow(
			ctx, q,
			int64(userID),
			string(secretType),
			comment,
			keyID,
			nonce,
			ciphertext,
			now,
		).Scan(&id)
	})
	if err != nil {
		return 0, fmt.Errorf("add secret: %w", err)
	}

	return domain.SecretID(id), nil
}

func (s *Store) ListSecrets(ctx context.Context, userID domain.UserID) ([]domain.SecretMeta, error) {
	const q = `
		SELECT id, type, comment, updated_at
		FROM secrets
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC;
	`

	var out []domain.SecretMeta

	err := withRetryRead(ctx, func() error {
		rows, err := s.pool.Query(ctx, q, int64(userID))
		if err != nil {
			return err
		}
		defer rows.Close()

		tmp := make([]domain.SecretMeta, 0, 16)

		for rows.Next() {
			var (
				id int64
				tp string
				m  domain.SecretMeta
			)
			if err := rows.Scan(&id, &tp, &m.Comment, &m.UpdatedAt); err != nil {
				return err
			}
			m.ID = domain.SecretID(id)
			m.Type = domain.SecretType(tp)
			tmp = append(tmp, m)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		out = tmp
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	return out, nil
}

func (s *Store) GetSecret(ctx context.Context, userID domain.UserID, secretID domain.SecretID) (domain.Secret, error) {
	const q = `
		SELECT id, type, comment, key_id, nonce, ciphertext, created_at, updated_at
		FROM secrets
		WHERE owner_id = $1 AND id = $2 AND deleted_at IS NULL;
	`

	var (
		sec domain.Secret
		id  int64
		tp  string
	)

	err := withRetryRead(ctx, func() error {
		return s.pool.QueryRow(ctx, q, int64(userID), int64(secretID)).Scan(
			&id,
			&tp,
			&sec.Comment,
			&sec.KeyID,
			&sec.Nonce,
			&sec.Ciphertext,
			&sec.CreatedAt,
			&sec.UpdatedAt,
		)
	})
	if err != nil {
		if isNoRows(err) {
			return domain.Secret{}, ErrSecretNotFound
		}
		return domain.Secret{}, fmt.Errorf("get secret: %w", err)
	}

	sec.ID = domain.SecretID(id)
	sec.OwnerID = userID
	sec.Type = domain.SecretType(tp)
	return sec, nil
}

func (s *Store) UpdateSecret(
	ctx context.Context,
	userID domain.UserID,
	secretID domain.SecretID,
	secretType domain.SecretType,
	comment string,
	keyID string,
	nonce []byte,
	ciphertext []byte,
) error {
	const q = `
		UPDATE secrets
		SET type = $1,
		    comment = $2,
		    key_id = $3,
		    nonce = $4,
		    ciphertext = $5,
		    updated_at = $6
		WHERE owner_id = $7 AND id = $8 AND deleted_at IS NULL;
	`

	now := time.Now().UTC()

	var affected int64
	err := withRetryWrite(ctx, func() error {
		ct, err := s.pool.Exec(ctx, q,
			string(secretType),
			comment,
			keyID,
			nonce,
			ciphertext,
			now,
			int64(userID),
			int64(secretID),
		)
		if err != nil {
			return err
		}
		affected = ct.RowsAffected()
		return nil
	})
	if err != nil {
		return fmt.Errorf("update secret: %w", err)
	}
	if affected == 0 {
		return ErrSecretNotFound
	}
	return nil
}

func (s *Store) DeleteSecret(ctx context.Context, userID domain.UserID, secretID domain.SecretID) error {
	const q = `
		UPDATE secrets
		SET deleted_at = now(), updated_at = now()
		WHERE owner_id = $1 AND id = $2 AND deleted_at IS NULL;
	`

	var affected int64
	err := withRetryWrite(ctx, func() error {
		ct, err := s.pool.Exec(ctx, q, int64(userID), int64(secretID))
		if err != nil {
			return err
		}
		affected = ct.RowsAffected()
		return nil
	})
	if err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}
	if affected == 0 {
		return ErrSecretNotFound
	}
	return nil
}
