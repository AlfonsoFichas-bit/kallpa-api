package chat

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"kallpa-api/internal/auth"
)

// Handler groups chat endpoints.
type Handler struct {
	service Service
}

// NewHandler creates a new chat Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ProcessMessage handles POST /v1/conversations/message
func (h *Handler) ProcessMessage(c *echo.Context) error {
	// Although the Telegram bot client calls this, it should still be authenticated
	// to ensure it belongs to the correct store and acts legitimately.
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	var req MessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request payload"})
	}
	if err := c.Echo().Validator.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "validation failed"})
	}

	// Security: Override StoreID with token StoreID to prevent cross-tenant impersonation
	req.StoreID = *claims.StoreID

	res, err := h.service.ProcessMessage(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error processing message"})
	}

	return c.JSON(http.StatusOK, res)
}

// RegisterHandlers mounts chat endpoints.
func RegisterHandlers(g *echo.Group, h *Handler) {
	g.POST("/conversations/message", h.ProcessMessage)
}
