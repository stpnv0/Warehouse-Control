package handler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   int
		wantMsg    string
	}{
		{"not found", domain.ErrNotFound, http.StatusNotFound, "not found"},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden, "insufficient permissions"},
		{"invalid credentials", domain.ErrInvalidCredentials, http.StatusUnauthorized, "invalid credentials"},
		{"invalid token", domain.ErrTokenInvalid, http.StatusUnauthorized, "invalid token"},
		{"token expired", domain.ErrTokenExpired, http.StatusUnauthorized, "token expired"},
		{"duplicate SKU", domain.ErrDuplicateSKU, http.StatusConflict, "item with this SKU already exists"},
		{"already exists", domain.ErrAlreadyExists, http.StatusConflict, "already exists"},
		{"no changes", domain.ErrNoChanges, http.StatusBadRequest, "no changes provided"},
		{"validation", domain.ErrValidation, http.StatusBadRequest, "validation error"},
		{"unknown", errors.New("something"), http.StatusInternalServerError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg := mapError(tt.err)
			assert.Equal(t, tt.wantCode, code)
			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}

func TestGetClaims_Missing(t *testing.T) {
	c, w := setupTestContext()

	result := getClaims(c)

	assert.Nil(t, result)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetClaims_WrongType(t *testing.T) {
	c, w := setupTestContext()
	c.Set("auth_claims", "not-a-claims-object")

	result := getClaims(c)

	assert.Nil(t, result)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetClaims_Success(t *testing.T) {
	c, _ := setupTestContext()
	setAuthClaims(c, testAdminClaims)

	result := getClaims(c)

	assert.NotNil(t, result)
	assert.Equal(t, testAdminClaims.UserID, result.UserID)
}
