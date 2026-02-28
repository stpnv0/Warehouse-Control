package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(hash)
}

func TestAuthService_Login_Success(t *testing.T) {
	userRepo := newMockuserRepository(t)
	tokenMgr := NewMockTokenManager(t)
	svc := NewAuthService(userRepo, tokenMgr, newTestLogger())

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: hashPassword(t, "password"),
		Role:         domain.RoleAdmin,
	}

	userRepo.EXPECT().GetByUsername(mock.Anything, "admin").Return(user, nil)
	tokenMgr.EXPECT().GenerateJWT(user).Return("jwt-token", nil)

	token, result, err := svc.Login(context.Background(), &domain.LoginInput{
		Username: "admin",
		Password: "password",
	})

	assert.NoError(t, err)
	assert.Equal(t, "jwt-token", token)
	assert.Equal(t, user.ID, result.ID)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := newMockuserRepository(t)
	tokenMgr := NewMockTokenManager(t)
	svc := NewAuthService(userRepo, tokenMgr, newTestLogger())

	userRepo.EXPECT().GetByUsername(mock.Anything, "unknown").Return(nil, domain.ErrNotFound)

	_, _, err := svc.Login(context.Background(), &domain.LoginInput{
		Username: "unknown",
		Password: "password",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	userRepo := newMockuserRepository(t)
	tokenMgr := NewMockTokenManager(t)
	svc := NewAuthService(userRepo, tokenMgr, newTestLogger())

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: hashPassword(t, "password"),
		Role:         domain.RoleAdmin,
	}

	userRepo.EXPECT().GetByUsername(mock.Anything, "admin").Return(user, nil)

	_, _, err := svc.Login(context.Background(), &domain.LoginInput{
		Username: "admin",
		Password: "wrong",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestAuthService_Login_RepoError(t *testing.T) {
	userRepo := newMockuserRepository(t)
	tokenMgr := NewMockTokenManager(t)
	svc := NewAuthService(userRepo, tokenMgr, newTestLogger())

	userRepo.EXPECT().GetByUsername(mock.Anything, "admin").Return(nil, errors.New("db error"))

	_, _, err := svc.Login(context.Background(), &domain.LoginInput{
		Username: "admin",
		Password: "password",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestAuthService_Login_TokenGenerationError(t *testing.T) {
	userRepo := newMockuserRepository(t)
	tokenMgr := NewMockTokenManager(t)
	svc := NewAuthService(userRepo, tokenMgr, newTestLogger())

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: hashPassword(t, "password"),
		Role:         domain.RoleAdmin,
	}

	userRepo.EXPECT().GetByUsername(mock.Anything, "admin").Return(user, nil)
	tokenMgr.EXPECT().GenerateJWT(user).Return("", errors.New("signing error"))

	_, _, err := svc.Login(context.Background(), &domain.LoginInput{
		Username: "admin",
		Password: "password",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generate token")
}

func TestAuthService_ListUsers_Success(t *testing.T) {
	userRepo := newMockuserRepository(t)
	tokenMgr := NewMockTokenManager(t)
	svc := NewAuthService(userRepo, tokenMgr, newTestLogger())

	users := []*domain.User{
		{ID: uuid.New(), Username: "admin", Role: domain.RoleAdmin},
		{ID: uuid.New(), Username: "viewer", Role: domain.RoleViewer},
	}

	userRepo.EXPECT().List(mock.Anything).Return(users, nil)

	result, err := svc.ListUsers(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestAuthService_ListUsers_Error(t *testing.T) {
	userRepo := newMockuserRepository(t)
	tokenMgr := NewMockTokenManager(t)
	svc := NewAuthService(userRepo, tokenMgr, newTestLogger())

	userRepo.EXPECT().List(mock.Anything).Return(nil, errors.New("db error"))

	_, err := svc.ListUsers(context.Background())

	assert.Error(t, err)
}
