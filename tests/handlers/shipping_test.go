package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/shipping"
)

// ShippingCustomValidator for echo
type ShippingCustomValidator struct {
	validator *validator.Validate
}

func (cv *ShippingCustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func TestListZones(t *testing.T) {
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/delivery-zones", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		expectedZones := []domain.DeliveryZone{
			{ID: 1, Name: "Downtown", Price: decimal.NewFromFloat(10.50)},
		}

		mockSvc.On("ListStoreZones", mock.Anything, storeID).Return(expectedZones, nil)

		err := h.ListZones(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res []domain.DeliveryZone
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, expectedZones[0].Name, res[0].Name)
	})

	t.Run("Unauthorized - No StoreID", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/delivery-zones", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		claims := &auth.CustomClaims{StoreID: nil}
		c.Set(auth.ContextKeyUser, claims)

		err := h.ListZones(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/delivery-zones", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("ListStoreZones", mock.Anything, storeID).Return(nil, errors.New("db error"))

		err := h.ListZones(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestAssignZone(t *testing.T) {
	e := echo.New()
	e.Validator = &ShippingCustomValidator{validator: validator.New()}

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		reqBody := shipping.AssignZoneRequest{ZoneID: 2}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/orders/1/shipping", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.POST("/orders/:id/shipping", h.AssignZone)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("AssignZoneToOrder", mock.Anything, storeID, int64(1), reqBody.ZoneID).Return(nil)

		err := h.AssignZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Invalid Order ID", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/orders/abc/shipping", bytes.NewReader([]byte("{}")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.POST("/orders/:id/shipping", h.AssignZone)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.AssignZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Zone Not Found", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		reqBody := shipping.AssignZoneRequest{ZoneID: 99}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/orders/1/shipping", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.POST("/orders/:id/shipping", h.AssignZone)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("AssignZoneToOrder", mock.Anything, storeID, int64(1), reqBody.ZoneID).Return(shipping.ErrZoneNotFound)

		err := h.AssignZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code) // ErrZoneNotFound mapped to 400
	})

	t.Run("Validation Failed", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		// Actually let's just make it empty JSON which omits ZoneID.

		req := httptest.NewRequest(http.MethodPost, "/orders/1/shipping", bytes.NewReader([]byte(`{}`)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.POST("/orders/:id/shipping", h.AssignZone)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.AssignZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestCreateZone(t *testing.T) {
	e := echo.New()
	e.Validator = &ShippingCustomValidator{validator: validator.New()}

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		reqBody := shipping.CreateZoneRequest{
			Name:  "North",
			Price: 15.0,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/delivery-zones", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		expectedZone := &domain.DeliveryZone{ID: 1, Name: "North", Price: decimal.NewFromFloat(15.0)}

		mockSvc.On("CreateZone", mock.Anything, storeID, reqBody.Name, decimal.NewFromFloat(reqBody.Price)).Return(expectedZone, nil)

		err := h.CreateZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("Validation Failed - Negative Price", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		reqBody := shipping.CreateZoneRequest{
			Name:  "North",
			Price: -5.0, // Fails min=0
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/delivery-zones", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.CreateZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		reqBody := shipping.CreateZoneRequest{
			Name:  "North",
			Price: 15.0,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/delivery-zones", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("CreateZone", mock.Anything, storeID, reqBody.Name, decimal.NewFromFloat(reqBody.Price)).Return(nil, errors.New("db err"))

		err := h.CreateZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestDeleteZone(t *testing.T) {
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		req := httptest.NewRequest(http.MethodDelete, "/delivery-zones/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.DELETE("/delivery-zones/:id", h.DeleteZone)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("DeleteZone", mock.Anything, storeID, int64(1)).Return(nil)

		err := h.DeleteZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		req := httptest.NewRequest(http.MethodDelete, "/delivery-zones/abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.DELETE("/delivery-zones/:id", h.DeleteZone)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.DeleteZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockSvc := new(MockShippingService)
		h := shipping.NewShippingHandler(mockSvc)

		req := httptest.NewRequest(http.MethodDelete, "/delivery-zones/999", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.DELETE("/delivery-zones/:id", h.DeleteZone)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("DeleteZone", mock.Anything, storeID, int64(999)).Return(shipping.ErrZoneNotFound)

		err := h.DeleteZone(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
