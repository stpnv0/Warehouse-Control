package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_GenerateAndValidate(t *testing.T) {
	m := NewManager("test-secret", time.Hour)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "admin",
		Role:     domain.RoleAdmin,
	}

	token, err := m.GenerateJWT(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := m.Validate(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	assert.Equal(t, domain.RoleAdmin, claims.Role)
}

func TestManager_ValidateExpiredToken(t *testing.T) {
	m := NewManager("test-secret", -time.Hour) // expired immediately

	user := &domain.User{
		ID:       uuid.New(),
		Username: "admin",
		Role:     domain.RoleAdmin,
	}

	token, err := m.GenerateJWT(user)
	require.NoError(t, err)

	_, err = m.Validate(token)
	assert.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestManager_ValidateInvalidToken(t *testing.T) {
	m := NewManager("test-secret", time.Hour)

	_, err := m.Validate("garbage-token")
	assert.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestManager_ValidateWrongSecret(t *testing.T) {
	m1 := NewManager("secret-1", time.Hour)
	m2 := NewManager("secret-2", time.Hour)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "admin",
		Role:     domain.RoleAdmin,
	}

	token, err := m1.GenerateJWT(user)
	require.NoError(t, err)

	_, err = m2.Validate(token)
	assert.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestManager_AllRoles(t *testing.T) {
	m := NewManager("test-secret", time.Hour)

	roles := []domain.Role{domain.RoleAdmin, domain.RoleManager, domain.RoleViewer}
	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			user := &domain.User{ID: uuid.New(), Username: string(role), Role: role}
			token, err := m.GenerateJWT(user)
			require.NoError(t, err)

			claims, err := m.Validate(token)
			require.NoError(t, err)
			assert.Equal(t, role, claims.Role)
		})
	}
}
