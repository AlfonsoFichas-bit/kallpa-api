package customers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"kallpa-api/internal/auth"
)

// CustomerUpsertRequest holds the payload for upserting a customer.
type CustomerUpsertRequest struct {
	TelegramID int64  `json:"telegram_id" validate:"required"`
	Name       string `json:"name" validate:"required"`
}

// CustomerHandler groups customer HTTP endpoints.
type CustomerHandler struct {
	service CustomerService
}

// NewCustomerHandler creates a new CustomerHandler.
func NewCustomerHandler(service CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

// ListCustomers handles GET /v1/customers
func (h *CustomerHandler) ListCustomers(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	status := c.QueryParam("status")
	customers, err := h.service.GetCustomers(c.Request().Context(), *claims.StoreID, status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error fetching customers"})
	}

	return c.JSON(http.StatusOK, customers)
}

// GetCustomer handles GET /v1/customers/:id
func (h *CustomerHandler) GetCustomer(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid customer id"})
	}

	customer, err := h.service.GetCustomer(c.Request().Context(), *claims.StoreID, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error fetching customer"})
	}
	if customer == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "customer not found"})
	}

	return c.JSON(http.StatusOK, customer)
}

// UpsertCustomer handles POST /v1/customers
func (h *CustomerHandler) UpsertCustomer(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil || claims.StoreID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "store context required"})
	}

	var req CustomerUpsertRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request"})
	}
	if err := c.Echo().Validator.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "validation failed"})
	}

	customer, err := h.service.UpsertCustomerFromTelegram(c.Request().Context(), *claims.StoreID, req.TelegramID, req.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error upserting customer"})
	}

	return c.JSON(http.StatusOK, customer)
}

// RegisterHandlers mounts customer routes to the given group.
func RegisterHandlers(g *echo.Group, h *CustomerHandler) {
	g.GET("/customers", h.ListCustomers)
	g.GET("/customers/:id", h.GetCustomer)
	g.POST("/customers", h.UpsertCustomer)
}
