package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/stores"
)

func TestGetStore(t *testing.T) {
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockStoresService)
		h := stores.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/stores/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/stores/:id", h.GetStore)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID, Role: domain.RoleStoreOwner}
		c.Set(auth.ContextKeyUser, claims)

		expectedStore := &domain.Store{ID: 1, Name: "Test Store"}
		mockSvc.On("GetStoreInfo", mock.Anything, int64(1)).Return(expectedStore, nil)

		err := h.GetStore(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res domain.Store
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, expectedStore.Name, res.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Success - Superadmin Accessing Any Store", func(t *testing.T) {
		mockSvc := new(MockStoresService)
		h := stores.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/stores/2", nil) // Requesting store 2
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/stores/:id", h.GetStore)
		e.Router().Route(c)

		storeID := int64(1) // Admin belongs to store 1
		claims := &auth.CustomClaims{StoreID: &storeID, Role: domain.RoleSuperadmin}
		c.Set(auth.ContextKeyUser, claims)

		expectedStore := &domain.Store{ID: 2, Name: "Another Store"}
		mockSvc.On("GetStoreInfo", mock.Anything, int64(2)).Return(expectedStore, nil)

		err := h.GetStore(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Unauthorized - No Claims", func(t *testing.T) {
		mockSvc := new(MockStoresService)
		h := stores.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/stores/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/stores/:id", h.GetStore)
		e.Router().Route(c)

		err := h.GetStore(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		mockSvc := new(MockStoresService)
		h := stores.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/stores/abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/stores/:id", h.GetStore)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID, Role: domain.RoleStoreOwner}
		c.Set(auth.ContextKeyUser, claims)

		err := h.GetStore(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Forbidden - Accessing Another Store", func(t *testing.T) {
		mockSvc := new(MockStoresService)
		h := stores.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/stores/2", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/stores/:id", h.GetStore)
		e.Router().Route(c)

		storeID := int64(1) // Admin belongs to store 1
		claims := &auth.CustomClaims{StoreID: &storeID, Role: domain.RoleStoreOwner}
		c.Set(auth.ContextKeyUser, claims)

		err := h.GetStore(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("Forbidden - No StoreID set for regular user", func(t *testing.T) {
		mockSvc := new(MockStoresService)
		h := stores.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/stores/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/stores/:id", h.GetStore)
		e.Router().Route(c)

		claims := &auth.CustomClaims{StoreID: nil, Role: domain.RoleStoreOwner}
		c.Set(auth.ContextKeyUser, claims)

		err := h.GetStore(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockSvc := new(MockStoresService)
		h := stores.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/stores/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/stores/:id", h.GetStore)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID, Role: domain.RoleStoreOwner}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("GetStoreInfo", mock.Anything, int64(1)).Return(nil, errors.New("store not found"))

		err := h.GetStore(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockStoresService)
		h := stores.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/stores/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/stores/:id", h.GetStore)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID, Role: domain.RoleStoreOwner}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("GetStoreInfo", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

		err := h.GetStore(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		mockSvc.AssertExpectations(t)
	})
}
