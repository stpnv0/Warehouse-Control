package handler

import (
	"errors"
	"net/http"

	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stpnv0/WarehouseControl/internal/middleware"
	"github.com/wb-go/wbf/ginext"
)

func writeJSON(c *ginext.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func writeError(c *ginext.Context, err error) {
	status, msg := mapError(err)
	c.JSON(status, ginext.H{"error": msg})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "insufficient permissions"
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid credentials"
	case errors.Is(err, domain.ErrTokenInvalid):
		return http.StatusUnauthorized, "invalid token"
	case errors.Is(err, domain.ErrTokenExpired):
		return http.StatusUnauthorized, "token expired"
	case errors.Is(err, domain.ErrDuplicateSKU):
		return http.StatusConflict, "item with this SKU already exists"
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict, "already exists"
	case errors.Is(err, domain.ErrNoChanges):
		return http.StatusBadRequest, "no changes provided"
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest, "validation error"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func getClaims(c *ginext.Context) *domain.AuthClaims {
	val, exists := c.Get(middleware.ClaimsKey)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ginext.H{
			"error": "unauthorized",
		})
		return nil
	}
	claims, ok := val.(*domain.AuthClaims)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{
			"error": "internal server error",
		})
		return nil
	}
	return claims
}
