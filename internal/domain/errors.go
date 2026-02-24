package domain

import "errors"

var (
	// Общие
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")

	// Авторизация
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("invalid token")

	// Права доступа
	ErrForbidden = errors.New("forbidden: insufficient permissions")

	// Валидация
	ErrValidation   = errors.New("validation error")
	ErrNoChanges    = errors.New("no changes provided")
	ErrDuplicateSKU = errors.New("item with this SKU already exists")
)
