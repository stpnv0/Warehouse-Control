package handler

import (
	"context"
	"net/http"

	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stpnv0/WarehouseControl/internal/handler/dto"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/logger"
)

type authService interface {
	Login(ctx context.Context, input *domain.LoginInput) (string, *domain.User, error)
	ListUsers(ctx context.Context) ([]*domain.User, error)
}

type AuthHandler struct {
	service authService
	log     logger.Logger
}

func NewAuthHandler(service authService, log logger.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		log:     log.With("handler", "auth"),
	}
}

// POST /api/auth/
func (h *AuthHandler) Login(c *ginext.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid request body"})
		return
	}

	token, user, err := h.service.Login(c.Request.Context(), req.ToInput())
	if err != nil {
		writeError(c, err)
		return
	}

	writeJSON(c, http.StatusOK, dto.NewLoginResponse(token, user))
}

// GET /api/auth/
func (h *AuthHandler) ListUsers(c *ginext.Context) {
	users, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}

	writeJSON(c, http.StatusOK, dto.NewUserListResponse(users))
}
