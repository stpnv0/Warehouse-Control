package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
)

type Manager struct {
	secret   []byte
	tokenTTL time.Duration
}

func NewManager(secret string, tokenTTL time.Duration) *Manager {
	return &Manager{
		secret:   []byte(secret),
		tokenTTL: tokenTTL,
	}
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func (m *Manager) GenerateJWT(user *domain.User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     string(user.Role),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Validate(tokenStr string) (*domain.AuthClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.secret, nil
		},
	)

	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrTokenInvalid
	}

	role, err := domain.ParseRole(claims.Role)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	return &domain.AuthClaims{
		UserID:   userID,
		Username: claims.Username,
		Role:     role,
	}, nil
}
