package domain

import (
	"strings"
	"time"
)

type UserID int64

type User struct {
	ID           UserID    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func ValidateCredentials(username, password string) error {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return ErrEmptyAuthData
	}
	return nil
}
