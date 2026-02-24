package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/wb-go/wbf/logger"
	"golang.org/x/crypto/bcrypt"
)

type userRepository interface {
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
}

type TokenManager interface {
	GenerateJWT(user *domain.User) (string, error)
	Validate(tokenStr string) (*domain.AuthClaims, error)
}

type AuthService struct {
	userRepo userRepository
	manager  TokenManager
	log      logger.Logger
}

func NewAuthService(userRepo userRepository, manager TokenManager, log logger.Logger) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		manager:  manager,
		log:      log.With("component", "AuthService"),
	}
}

func (s *AuthService) Login(ctx context.Context, input *domain.LoginInput) (string, *domain.User, error) {
	const op = "AuthService.Login"
	user, err := s.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", nil, domain.ErrInvalidCredentials
		}
		s.log.Ctx(ctx).Error("failed to get user",
			"error", err,
			"username", input.Username,
		)
		return "", nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := s.manager.GenerateJWT(user)
	if err != nil {
		s.log.Ctx(ctx).Error("failed to generate token",
			"error", err,
			"user_id", user.ID,
		)
		return "", nil, fmt.Errorf("%s - generate token: %w", op, err)
	}

	return token, user, nil
}

func (s *AuthService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	const op = "AuthService.ListUsers"

	users, err := s.userRepo.List(ctx)
	if err != nil {
		s.log.Ctx(ctx).Error("failed to list users",
			"error", err,
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

//TODO: добавить методы Register, Logout, однако это по заданию не требуется
