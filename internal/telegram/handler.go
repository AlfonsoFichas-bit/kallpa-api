package telegram

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Handler groups Telegram webhook endpoints.
type Handler struct {
	service Service
}

// NewHandler creates a new Telegram webhook Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// HandleWebhook processes POST /v1/webhook/telegram/:token
func (h *Handler) HandleWebhook(c *echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "bot token required"})
	}

	var update Update
	if err := c.Bind(&update); err != nil {
		// Telegram expects a 200 OK even if we can't parse it, otherwise it retries indefinitely.
		return c.JSON(http.StatusOK, map[string]string{"message": "ignored"})
	}

	// Process the webhook asynchronously or synchronously.
	// Telegram expects quick responses, but for simplicity we'll do it synchronously here.
	_ = h.service.ProcessWebhook(c.Request().Context(), token, &update)

	return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
}

// RegisterHandlers mounts webhook endpoints.
func RegisterHandlers(g *echo.Group, h *Handler) {
	g.POST("/webhook/telegram/:token", h.HandleWebhook)
}
