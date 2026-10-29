package dashboard

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"kallpa-api/internal/auth"
)

// Handler groups dashboard endpoints.
type Handler struct {
	service Service
}

// NewHandler creates a new dashboard Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetMetrics handles GET /v1/dashboard/metrics
func (h *Handler) GetMetrics(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	metrics, err := h.service.GetDashboardMetrics(c.Request().Context(), *claims.StoreID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error fetching metrics"})
	}

	return c.JSON(http.StatusOK, metrics)
}

// RegisterHandlers mounts dashboard routes.
func RegisterHandlers(g *echo.Group, h *Handler) {
	g.GET("/dashboard/metrics", h.GetMetrics)
}
