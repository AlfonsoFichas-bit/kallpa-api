package shipping

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"kallpa-api/internal/auth"
)

// AssignZoneRequest is the payload to assign shipping to an order.
type AssignZoneRequest struct {
	ZoneID int64 `json:"zone_id" validate:"required"`
}

// ShippingHandler groups HTTP endpoints for shipping.
type ShippingHandler struct {
	service ShippingService
}

// NewShippingHandler creates a new ShippingHandler.
func NewShippingHandler(service ShippingService) *ShippingHandler {
	return &ShippingHandler{service: service}
}

// ListZones handles GET /v1/delivery-zones
func (h *ShippingHandler) ListZones(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	zones, err := h.service.ListStoreZones(c.Request().Context(), *claims.StoreID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error fetching zones"})
	}

	return c.JSON(http.StatusOK, zones)
}

// AssignZone handles POST /v1/orders/:id/shipping
func (h *ShippingHandler) AssignZone(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid order id"})
	}

	var req AssignZoneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request"})
	}
	if err := c.Echo().Validator.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "validation failed"})
	}

	err = h.service.AssignZoneToOrder(c.Request().Context(), *claims.StoreID, orderID, req.ZoneID)
	if err != nil {
		if errors.Is(err, ErrZoneNotFound) {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error assigning zone"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "shipping assigned successfully"})
}

// RegisterHandlers mounts shipping routes to the given group.
func RegisterHandlers(g *echo.Group, h *ShippingHandler) {
	g.GET("/delivery-zones", h.ListZones)
	g.POST("/orders/:id/shipping", h.AssignZone)
}
