package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuth_NoHeader(t *testing.T) {
	validator := NewMockTokenValidator(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	handler := Auth(validator)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestAuth_EmptyBearerToken(t *testing.T) {
	validator := NewMockTokenValidator(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer ")

	handler := Auth(validator)
	handler(c)

	// extractBearerToken trims spaces; empty token passed to validator
	// The validator won't be called since token is empty after trim
	// Actually let me check - "Bearer " split => ["Bearer", ""] => TrimSpace("") => ""
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	validator := NewMockTokenValidator(t)
	validator.EXPECT().Validate("bad-token").Return(nil, domain.ErrTokenInvalid)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer bad-token")

	handler := Auth(validator)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestAuth_ValidToken(t *testing.T) {
	validator := NewMockTokenValidator(t)
	claims := &domain.AuthClaims{
		UserID:   uuid.New(),
		Username: "admin",
		Role:     domain.RoleAdmin,
	}
	validator.EXPECT().Validate("valid-token").Return(claims, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	handler := Auth(validator)
	handler(c)

	assert.False(t, c.IsAborted())

	val, exists := c.Get(ClaimsKey)
	assert.True(t, exists)
	assert.Equal(t, claims, val)
}

func TestAuth_NonBearerScheme(t *testing.T) {
	validator := NewMockTokenValidator(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

	handler := Auth(validator)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"valid", "Bearer token123", "token123"},
		{"case insensitive", "bearer token123", "token123"},
		{"no header", "", ""},
		{"no bearer prefix", "token123", ""},
		{"basic auth", "Basic abc", ""},
		{"bearer only", "Bearer", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.header != "" {
				c.Request.Header.Set("Authorization", tt.header)
			}

			result := extractBearerToken(c)
			assert.Equal(t, tt.want, result)
		})
	}
}
