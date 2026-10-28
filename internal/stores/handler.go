package stores

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/domain"
)

// Handler groups store endpoints.
type Handler struct {
	service Service
}

// NewHandler creates a new store Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetStore handles GET /v1/stores/:id
func (h *Handler) GetStore(c *echo.Context) error {
	claims := auth.GetUserClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid store id"})
	}

	// Security: users (except superadmin) can only query their own store
	if claims.Role != domain.RoleSuperadmin {
		if claims.StoreID == nil || *claims.StoreID != id {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "forbidden to access this store"})
		}
	}

	store, err := h.service.GetStoreInfo(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "store not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error fetching store"})
	}

	return c.JSON(http.StatusOK, store)
}

// RegisterHandlers mounts store endpoints.
func RegisterHandlers(g *echo.Group, h *Handler) {
	g.GET("/stores/:id", h.GetStore)
}
