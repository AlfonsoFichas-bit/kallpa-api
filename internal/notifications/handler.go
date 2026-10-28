package notifications

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"kallpa-api/internal/auth"
)

// Handler groups notification endpoints.
type Handler struct {
	service Service
}

// NewHandler creates a new notification Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ListNotifications handles GET /v1/notifications
func (h *Handler) ListNotifications(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	unreadOnly := c.QueryParam("unread") == "true"

	notifs, err := h.service.GetNotifications(c.Request().Context(), *claims.StoreID, unreadOnly)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error fetching notifications"})
	}

	return c.JSON(http.StatusOK, notifs)
}

// MarkAsRead handles PATCH /v1/notifications/:id/read
func (h *Handler) MarkAsRead(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid notification id"})
	}

	if err := h.service.MarkNotificationAsRead(c.Request().Context(), *claims.StoreID, id); err != nil {
		if err.Error() == "notification not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error updating notification"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "notification marked as read"})
}

// RegisterHandlers mounts notification endpoints.
func RegisterHandlers(g *echo.Group, h *Handler) {
	g.GET("/notifications", h.ListNotifications)
	g.PATCH("/notifications/:id/read", h.MarkAsRead)
}
