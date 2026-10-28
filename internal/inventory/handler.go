package inventory

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"kallpa-api/internal/auth"
)

// StockMovementRequest holds the payload for updating stock.
type StockMovementRequest struct {
	VariantID int64 `json:"variant_id" validate:"required"`
	Delta     int   `json:"delta" validate:"required"`
}

// InventoryHandler groups inventory HTTP endpoints.
type InventoryHandler struct {
	service InventoryService
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler(service InventoryService) *InventoryHandler {
	return &InventoryHandler{service: service}
}

// ListProducts handles GET /v1/products
func (h *InventoryHandler) ListProducts(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	search := c.QueryParam("search")
	products, err := h.service.SearchProducts(c.Request().Context(), *claims.StoreID, search)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error fetching products"})
	}

	return c.JSON(http.StatusOK, products)
}

// GetProduct handles GET /v1/products/:id
func (h *InventoryHandler) GetProduct(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid product id"})
	}

	product, err := h.service.GetProduct(c.Request().Context(), *claims.StoreID, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error fetching product"})
	}
	if product == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "product not found"})
	}

	return c.JSON(http.StatusOK, product)
}

// AdjustStock handles POST /v1/stock-movements
func (h *InventoryHandler) AdjustStock(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	var req StockMovementRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request"})
	}
	if err := c.Echo().Validator.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "validation failed"})
	}

	err := h.service.AdjustStock(c.Request().Context(), *claims.StoreID, req.VariantID, req.Delta)
	if err != nil {
		if errors.Is(err, ErrInsufficientStock) {
			return c.JSON(http.StatusConflict, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "stock updated successfully"})
}

// GetLowStock handles GET /v1/variants/low-stock
func (h *InventoryHandler) GetLowStock(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	thresholdStr := c.QueryParam("threshold")
	threshold := 5 // default
	if thresholdStr != "" {
		if t, err := strconv.Atoi(thresholdStr); err == nil {
			threshold = t
		}
	}

	variants, err := h.service.GetLowStock(c.Request().Context(), *claims.StoreID, threshold)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error fetching low stock"})
	}

	return c.JSON(http.StatusOK, variants)
}

// RegisterHandlers mounts inventory routes to the given group.
func RegisterHandlers(g *echo.Group, h *InventoryHandler) {
	g.GET("/products", h.ListProducts)
	g.GET("/products/:id", h.GetProduct)
	g.POST("/stock-movements", h.AdjustStock)
	g.GET("/variants/low-stock", h.GetLowStock)
}
