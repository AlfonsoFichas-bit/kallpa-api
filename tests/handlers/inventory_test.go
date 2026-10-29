package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
)

func TestInventoryHandler_ListProducts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockInventoryService)
		handler := inventory.NewInventoryHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/products?search=shirt", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		expectedProducts := []domain.Product{
			{ID: 1, Name: "Shirt"},
		}
		mockService.On("SearchProducts", mock.Anything, storeID, "shirt").Return(expectedProducts, nil)

		err := handler.ListProducts(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res []domain.Product
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Len(t, res, 1)

		mockService.AssertExpectations(t)
	})
}

func TestInventoryHandler_GetProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockInventoryService)
		handler := inventory.NewInventoryHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/products/:id", handler.GetProduct)
		e.Router().Route(c)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		expectedProduct := &domain.Product{ID: 1, Name: "Shirt"}
		mockService.On("GetProduct", mock.Anything, storeID, int64(1)).Return(expectedProduct, nil)

		err := handler.GetProduct(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockService.AssertExpectations(t)
	})
}

func TestInventoryHandler_AdjustStock(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockInventoryService)
		handler := inventory.NewInventoryHandler(mockService)

		reqBody := inventory.StockMovementRequest{VariantID: 5, Delta: 10}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/stock-movements", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		mockService.On("AdjustStock", mock.Anything, storeID, int64(5), 10).Return(nil)

		err := handler.AdjustStock(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("insufficient_stock", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockInventoryService)
		handler := inventory.NewInventoryHandler(mockService)

		reqBody := inventory.StockMovementRequest{VariantID: 5, Delta: -100}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/stock-movements", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		mockService.On("AdjustStock", mock.Anything, storeID, int64(5), -100).Return(inventory.ErrInsufficientStock)

		err := handler.AdjustStock(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})
}

func TestInventoryHandler_GetLowStock(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockInventoryService)
		handler := inventory.NewInventoryHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/variants/low-stock?threshold=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		expectedVariants := []domain.ProductVariant{{ID: 5, StockQuantity: 2}}
		mockService.On("GetLowStock", mock.Anything, storeID, 10).Return(expectedVariants, nil)

		err := handler.GetLowStock(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockService.AssertExpectations(t)
	})
}

func TestInventoryHandler_ToggleProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockInventoryService)
		handler := inventory.NewInventoryHandler(mockService)

		req := httptest.NewRequest(http.MethodPatch, "/products/1/toggle", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.PATCH("/products/:id/toggle", handler.ToggleProduct)
		e.Router().Route(c)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		mockService.On("ToggleProductStatus", mock.Anything, storeID, int64(1)).Return(nil)

		err := handler.ToggleProduct(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockService.AssertExpectations(t)
	})
}
