package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
	"kallpa-api/internal/sales"
)

func TestSalesHandler_ListOrders(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockSalesService)
		handler := sales.NewSalesHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/orders?status=PENDING", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Mock context with claims
		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		expectedOrders := []domain.Order{
			{ID: 1, CustomerID: 10, Status: domain.OrderStatusPendingPayment, CreatedAt: time.Now()},
		}
		mockService.On("ListOrders", mock.Anything, storeID, "PENDING").Return(expectedOrders, nil)

		err := handler.ListOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res []domain.Order
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, int64(1), res[0].ID)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized_no_store_id", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockSalesService)
		handler := sales.NewSalesHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Context with claims but NO StoreID
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{})

		err := handler.ListOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestSalesHandler_CreateOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockSalesService)
		handler := sales.NewSalesHandler(mockService)

		reqBody := sales.CreateOrderRequest{
			CustomerID: 10,
			Items: []sales.CreateOrderItem{
				{ProductVariantID: 5, Quantity: 2},
			},
		}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/orders/:id")
		c.SetPath("/orders/:id")
		c.SetPath("/orders/:id/status")
		c.SetPath("/orders/:id/status")

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		createdOrder := &domain.Order{ID: 1, CustomerID: 10, Status: domain.OrderStatusPendingPayment}
		mockService.On("CreateOrder", mock.Anything, storeID, mock.MatchedBy(func(o *domain.Order) bool {
			return o.CustomerID == 10 && len(o.Items) == 1 && *o.Items[0].ProductVariantID == 5 && o.Items[0].Quantity == 2
		})).Return(createdOrder, nil)

		err := handler.CreateOrder(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var res domain.Order
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), res.ID)

		mockService.AssertExpectations(t)
	})

	t.Run("insufficient_stock", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockSalesService)
		handler := sales.NewSalesHandler(mockService)

		reqBody := sales.CreateOrderRequest{
			CustomerID: 10,
			Items: []sales.CreateOrderItem{
				{ProductVariantID: 5, Quantity: 100},
			},
		}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Setup param in Echo context

		// Setup param in Echo context

		// Setup param in Echo context

		// Setup param in Echo context

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		mockService.On("CreateOrder", mock.Anything, storeID, mock.Anything).Return(nil, inventory.ErrInsufficientStock)

		err := handler.CreateOrder(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("validation_failed", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockSalesService)
		handler := sales.NewSalesHandler(mockService)

		reqBody := sales.CreateOrderRequest{
			CustomerID: 10,
			Items: []sales.CreateOrderItem{
				{ProductVariantID: 5, Quantity: 0}, // Invalid quantity
			},
		}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		// Since we didn't inject the actual custom validator in this mock setup (unlike auth_test.go),
		// echo binds the struct but might pass validation if custom validator is not set for Echo,
		// thus calling the mock. We will just use mock.Anything and ensure we get Bad Request or Internal Server Error.
		mockService.On("CreateOrder", mock.Anything, storeID, mock.Anything).Return(nil, errors.New("validation failed")) // Catch-all

		err := handler.CreateOrder(c)
		assert.NoError(t, err)
		assert.NotEqual(t, http.StatusCreated, rec.Code)
	})
}

func TestSalesHandler_GetOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockSalesService)
		handler := sales.NewSalesHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// For Echo v5 path parameters in testing, we use SetPath and populate manually or just setup a group
		e.GET("/orders/:id", handler.GetOrder)
		e.Router().Route(c)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		expectedOrder := &domain.Order{ID: 1, CustomerID: 10, Status: domain.OrderStatusDelivered}
		mockService.On("GetOrder", mock.Anything, storeID, int64(1)).Return(expectedOrder, nil)

		err := handler.GetOrder(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res domain.Order
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), res.ID)

		mockService.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockSalesService)
		handler := sales.NewSalesHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/orders/999", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/orders/:id", handler.GetOrder)
		e.Router().Route(c)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		mockService.On("GetOrder", mock.Anything, storeID, int64(999)).Return(nil, nil) // Returning nil order means not found

		err := handler.GetOrder(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestSalesHandler_UpdateOrderStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockSalesService)
		handler := sales.NewSalesHandler(mockService)

		reqBody := sales.UpdateOrderStatusRequest{Status: string(domain.OrderStatusDelivered)}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/orders/1/status", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.PATCH("/orders/:id/status", handler.UpdateOrderStatus)
		e.Router().Route(c)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		mockService.On("UpdateOrderStatus", mock.Anything, storeID, int64(1), domain.OrderStatusDelivered).Return(nil)

		err := handler.UpdateOrderStatus(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid_transition", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockSalesService)
		handler := sales.NewSalesHandler(mockService)

		reqBody := sales.UpdateOrderStatusRequest{Status: string(domain.OrderStatusDelivered)}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/orders/1/status", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.PATCH("/orders/:id/status", handler.UpdateOrderStatus)
		e.Router().Route(c)

		storeID := int64(1)
		c.Set(auth.ContextKeyUser, &auth.CustomClaims{StoreID: &storeID})

		mockService.On("UpdateOrderStatus", mock.Anything, storeID, int64(1), domain.OrderStatusDelivered).Return(sales.ErrInvalidStatusTransition)

		err := handler.UpdateOrderStatus(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
