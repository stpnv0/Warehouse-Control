package dto

import (
	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
)

// DTO для POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (r *LoginRequest) ToInput() *domain.LoginInput {
	return &domain.LoginInput{
		Username: r.Username,
		Password: r.Password,
	}
}

// UserResponse - публичное представление пользователя (без пароля)
type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
}

func NewUserResponse(u *domain.User) *UserResponse {
	return &UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Role:     string(u.Role),
	}
}

func NewUserListResponse(users []*domain.User) []*UserResponse {
	resp := make([]*UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, NewUserResponse(u))
	}
	return resp
}

// DTO ответа на авторизацию
type LoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

func NewLoginResponse(token string, user *domain.User) *LoginResponse {
	return &LoginResponse{
		Token: token,
		User:  NewUserResponse(user),
	}
}
