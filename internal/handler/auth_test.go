package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stpnv0/WarehouseControl/internal/handler/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_Login_Success(t *testing.T) {
	svc := newMockauthService(t)
	h := NewAuthHandler(svc, newTestLogger())

	user := &domain.User{
		ID:       uuid.New(),
		Username: "admin",
		Role:     domain.RoleAdmin,
	}

	svc.EXPECT().Login(mock.Anything, &domain.LoginInput{
		Username: "admin",
		Password: "password",
	}).Return("jwt-token", user, nil)

	body, _ := json.Marshal(dto.LoginRequest{Username: "admin", Password: "password"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.LoginResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "jwt-token", resp.Token)
	assert.Equal(t, "admin", resp.User.Username)
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	svc := newMockauthService(t)
	h := NewAuthHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(`{}`)))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	svc := newMockauthService(t)
	h := NewAuthHandler(svc, newTestLogger())

	svc.EXPECT().Login(mock.Anything, mock.Anything).Return("", nil, domain.ErrInvalidCredentials)

	body, _ := json.Marshal(dto.LoginRequest{Username: "admin", Password: "wrong"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_ListUsers_Success(t *testing.T) {
	svc := newMockauthService(t)
	h := NewAuthHandler(svc, newTestLogger())

	users := []*domain.User{
		{ID: uuid.New(), Username: "admin", Role: domain.RoleAdmin},
		{ID: uuid.New(), Username: "viewer", Role: domain.RoleViewer},
	}

	svc.EXPECT().ListUsers(mock.Anything).Return(users, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/users", nil)

	h.ListUsers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []*dto.UserResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}
