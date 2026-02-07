package postgres

import (
	"errors"
	"yandex-gophkeeper/internal/domain"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

var (
	ErrUserAlreadyExists = domain.ErrLoginAlreadyTaken
	ErrUserNotFound      = domain.ErrUserNotFound
	ErrSecretNotFound    = domain.ErrSecretNotFound
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
