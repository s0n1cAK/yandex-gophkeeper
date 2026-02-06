package domain

import "errors"

var (
	ErrEmptyUsername     = errors.New("username is empty")
	ErrEmptyPassword     = errors.New("password is empty")
	ErrEmptyAuthData     = errors.New("username and password cannot be empty")
	ErrLoginAlreadyTaken = errors.New("login already taken")
	ErrUserNotFound      = errors.New("user not found")
	ErrUnauthorized      = errors.New("unauthorized")

	ErrInvalidSecretType = errors.New("secret type is invalid")
	ErrSecretNotFound    = errors.New("secret not found")

	ErrEmptyCardNumber         = errors.New("card number is empty")
	ErrInvalidCardNumberFormat = errors.New("card number has invalid format")
	ErrInvalidBankCardNumber   = errors.New("card number is invalid")
	ErrEmptyCardExpiry         = errors.New("card expiry is empty")

	ErrInvalidPayload = errors.New("invalid payload")
	ErrInternal       = errors.New("internal error")
)
