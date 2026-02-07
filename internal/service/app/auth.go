package app

import (
	"context"
	"errors"
	"strings"

	"yandex-gophkeeper/internal/domain"
	postgres "yandex-gophkeeper/internal/storage/postgresql"

	"golang.org/x/crypto/bcrypt"
)

type UsersStore interface {
	CreateUser(ctx context.Context, username, passwordHash string) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
}

type AuthService struct {
	store  UsersStore
	tokens *TokenManager
}

func NewAuthService(store UsersStore, tokens *TokenManager) *AuthService {
	return &AuthService{store: store, tokens: tokens}
}

func (a *AuthService) Register(ctx context.Context, username, password string) (string, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if err := domain.ValidateCredentials(username, password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	u, err := a.store.CreateUser(ctx, username, string(hash))
	if err != nil {
		if errors.Is(err, postgres.ErrUserAlreadyExists) {
			return "", err
		}
		return "", err
	}

	return a.tokens.Issue(u.ID)
}

func (a *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if err := domain.ValidateCredentials(username, password); err != nil {
		return "", err
	}

	u, err := a.store.GetUserByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", err
	}

	return a.tokens.Issue(u.ID)
}
