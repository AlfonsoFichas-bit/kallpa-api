package auth

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

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
		if err == ErrInvalidCredentials {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "internal server error"})
	}

	return c.JSON(http.StatusOK, tokens)
}

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

func RegisterHandlers(g *echo.Group, handler *AuthHandler) {
	g.POST("/login", handler.Login)
	g.POST("/refresh", handler.Refresh)
}
