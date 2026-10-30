package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/dashboard"
)

func TestGetMetrics(t *testing.T) {
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockDashboardService)
		h := dashboard.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		expectedRes := &dashboard.MetricsResponse{
			TotalIncome:   decimal.NewFromFloat(150.50),
			PendingOrders: 5,
		}

		mockSvc.On("GetDashboardMetrics", mock.Anything, storeID).Return(expectedRes, nil)

		err := h.GetMetrics(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res dashboard.MetricsResponse
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, expectedRes.PendingOrders, res.PendingOrders)
		assert.True(t, expectedRes.TotalIncome.Equal(res.TotalIncome))

		mockSvc.AssertExpectations(t)
	})

	t.Run("Unauthorized - No Claims", func(t *testing.T) {
		mockSvc := new(MockDashboardService)
		h := dashboard.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.GetMetrics(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Unauthorized - No StoreID", func(t *testing.T) {
		mockSvc := new(MockDashboardService)
		h := dashboard.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		claims := &auth.CustomClaims{StoreID: nil}
		c.Set(auth.ContextKeyUser, claims)

		err := h.GetMetrics(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockDashboardService)
		h := dashboard.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("GetDashboardMetrics", mock.Anything, storeID).Return(nil, errors.New("db error"))

		err := h.GetMetrics(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		mockSvc.AssertExpectations(t)
	})
}
