package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"kallpa-api/internal/domain"
)

// ContextKeyUser is the key used to store claims in Echo context.
const ContextKeyUser = "user_claims"

// AuthMiddleware creates a middleware to extract and validate JWT from Authorization header.
func AuthMiddleware(authService AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "missing authorization header"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "invalid authorization header format"})
			}

			tokenStr := parts[1]
			claims, err := authService.ValidateToken(tokenStr)
			if err != nil || claims.Type != "access" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "invalid or expired token"})
			}

			// Inject claims into context
			c.Set(ContextKeyUser, claims)

			return next(c)
		}
	}
}

// RequireRole creates a middleware helper to enforce specific roles.
func RequireRole(allowedRoles ...domain.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			claims, ok := c.Get(ContextKeyUser).(*CustomClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "user context not found"})
			}

			isAllowed := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				return c.JSON(http.StatusForbidden, ErrorResponse{Message: "insufficient role"})
			}

			return next(c)
		}
	}
}

// GetUserClaims retrieves the user claims extracted from context.
func GetUserClaims(c *echo.Context) *CustomClaims {
	claims, _ := c.Get(ContextKeyUser).(*CustomClaims)
	return claims
}
