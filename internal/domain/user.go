package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"         db:"id"`
	Username     string    `json:"username"    db:"username"`
	PasswordHash string    `json:"-"           db:"password_hash"`
	Role         Role      `json:"role"        db:"role"`
	CreatedAt    time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"  db:"updated_at"`
}

// LoginInput - входные данные для авторизации
type LoginInput struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthClaims - данные, зашиваемые в JWT
type AuthClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     Role      `json:"role"`
}
