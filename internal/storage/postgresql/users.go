package postgres

import (
	"context"
	"fmt"
	"time"
	"yandex-gophkeeper/internal/domain"
)

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) (domain.User, error) {
	const q = `
		INSERT INTO users(username, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`

	now := time.Now().UTC()

	var (
		id        int64
		createdAt time.Time
	)

	err := withRetryWrite(ctx, func() error {
		return s.pool.QueryRow(ctx, q, username, passwordHash, now).Scan(&id, &createdAt)
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, ErrUserAlreadyExists
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return domain.User{
		ID:           domain.UserID(id),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
	}, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	const q = `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = $1;
	`

	var (
		u  domain.User
		id int64
	)

	err := withRetryRead(ctx, func() error {
		return s.pool.QueryRow(ctx, q, username).Scan(&id, &u.Username, &u.PasswordHash, &u.CreatedAt)
	})
	if err != nil {
		if isNoRows(err) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("get user by username: %w", err)
	}

	u.ID = domain.UserID(id)
	return u, nil
}
