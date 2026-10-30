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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/customers"
	"kallpa-api/internal/domain"
)

// CustomersCustomValidator for echo
type CustomersCustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomersCustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func TestListCustomers(t *testing.T) {
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/customers?status=active", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		fullName := "Test Customer"
		expectedCustomers := []domain.Customer{
			{ID: 1, FullName: &fullName},
		}

		mockSvc.On("GetCustomers", mock.Anything, storeID, "active").Return(expectedCustomers, nil)

		err := h.ListCustomers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res []domain.Customer
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, expectedCustomers[0].FullName, res[0].FullName)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Unauthorized - No Claims", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/customers", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.ListCustomers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/customers", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("GetCustomers", mock.Anything, storeID, "").Return(nil, errors.New("db error"))

		err := h.ListCustomers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetCustomer(t *testing.T) {
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/customers/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/customers/:id", h.GetCustomer)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		fullName := "Test Customer"
		expectedCustomer := &domain.Customer{ID: 1, FullName: &fullName}
		mockSvc.On("GetCustomer", mock.Anything, storeID, int64(1)).Return(expectedCustomer, nil)

		err := h.GetCustomer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res domain.Customer
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer.FullName, res.FullName)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/customers/abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/customers/:id", h.GetCustomer)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.GetCustomer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/customers/999", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/customers/:id", h.GetCustomer)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("GetCustomer", mock.Anything, storeID, int64(999)).Return(nil, nil)

		err := h.GetCustomer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/customers/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.GET("/customers/:id", h.GetCustomer)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("GetCustomer", mock.Anything, storeID, int64(1)).Return(nil, errors.New("db err"))

		err := h.GetCustomer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestUpsertCustomer(t *testing.T) {
	e := echo.New()
	e.Validator = &CustomersCustomValidator{validator: validator.New()}

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		reqBody := customers.CustomerUpsertRequest{
			TelegramID: 12345,
			Name:       "Test User",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		fullName := "Test User"
		expectedCustomer := &domain.Customer{ID: 1, FullName: &fullName}
		mockSvc.On("UpsertCustomerFromTelegram", mock.Anything, storeID, reqBody.TelegramID, reqBody.Name).Return(expectedCustomer, nil)

		err := h.UpsertCustomer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Validation Failed - Missing Required Field", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		reqBody := customers.CustomerUpsertRequest{
			Name: "Test User", // Missing TelegramID (which will be 0 and fail required validation depending on setup, but let's say Name is empty)
		}
		reqBody.Name = ""
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.UpsertCustomer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader([]byte("{invalid json}")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.UpsertCustomer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockCustomerService)
		h := customers.NewCustomerHandler(mockSvc)

		reqBody := customers.CustomerUpsertRequest{
			TelegramID: 12345,
			Name:       "Test User",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("UpsertCustomerFromTelegram", mock.Anything, storeID, reqBody.TelegramID, reqBody.Name).Return(nil, errors.New("db err"))

		err := h.UpsertCustomer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
