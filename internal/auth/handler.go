package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
)

// LoginRequest holds the expected payload for the login endpoint.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest holds the expected payload for the refresh token endpoint.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ErrorResponse represents standard error structures.
type ErrorResponse struct {
	Message string `json:"message"`
}

// AuthHandler groups authentication routes.
type AuthHandler struct {
	service AuthService
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login handles user authentication and returns JWT tokens.
func (h *AuthHandler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request payload"})
	}

	if err := c.Echo().Validator.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "validation failed"})
	}

	tokens, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "internal server error"})
	}

	return c.JSON(http.StatusOK, tokens)
}

// Refresh handles token renewal using a valid refresh token.
func (h *AuthHandler) Refresh(c *echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request payload"})
	}

	if err := c.Echo().Validator.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "validation failed"})
	}

	tokens, err := h.service.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "invalid or expired refresh token"})
	}

	return c.JSON(http.StatusOK, tokens)
}

// RegisterHandlers mounts auth endpoints onto the provided Echo group.
func RegisterHandlers(g *echo.Group, handler *AuthHandler) {
	g.POST("/login", handler.Login)
	g.POST("/refresh", handler.Refresh)
}
