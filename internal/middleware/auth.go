package middleware

import (
	"net/http"
	"strings"

	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/wb-go/wbf/ginext"
)

const ClaimsKey = "auth_claims"

type TokenValidator interface {
	Validate(tokenStr string) (*domain.AuthClaims, error)
}

func Auth(validator TokenValidator) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginext.H{
				"error": "authorization token required",
			})
			return
		}

		claims, err := validator.Validate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginext.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

func extractBearerToken(c *ginext.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
