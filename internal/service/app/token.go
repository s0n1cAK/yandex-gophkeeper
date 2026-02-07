package app

import (
	"errors"
	"time"
	"yandex-gophkeeper/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

var ErrEmptyJWTSecret = errors.New("service: empty jwt secret")

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

type Claims struct {
	UID int64 `json:"uid"`
	jwt.RegisteredClaims
}

func NewTokenManager(jwtSecret string, ttl time.Duration) (*TokenManager, error) {
	if jwtSecret == "" {
		return nil, ErrEmptyJWTSecret
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &TokenManager{secret: []byte(jwtSecret), ttl: ttl}, nil
}

func (m *TokenManager) Issue(userID domain.UserID) (string, error) {
	now := time.Now()
	claims := Claims{
		UID: int64(userID),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

func (m *TokenManager) Verify(token string) (int64, bool) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	})
	if err != nil || !parsed.Valid {
		return 0, false
	}

	c, ok := parsed.Claims.(*Claims)
	if !ok {
		return 0, false
	}
	return c.UID, true
}
