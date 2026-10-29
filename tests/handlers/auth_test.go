package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/auth"
)

// CustomValidator for tests
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	return e
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockAuthService)
		handler := auth.NewAuthHandler(mockService)

		reqBody := auth.LoginRequest{Email: "test@example.com", Password: "password123"}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := &auth.TokenResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
		}

		mockService.On("Login", mock.Anything, "test@example.com", "password123").Return(expectedResponse, nil)

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res auth.TokenResponse
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.AccessToken, res.AccessToken)
		assert.Equal(t, expectedResponse.RefreshToken, res.RefreshToken)

		mockService.AssertExpectations(t)
	})

	t.Run("validation_failed", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockAuthService)
		handler := auth.NewAuthHandler(mockService)

		reqBody := auth.LoginRequest{Email: "invalid_email"} // Missing password and invalid email
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var res auth.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, "validation failed", res.Message)
	})

	t.Run("invalid_credentials", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockAuthService)
		handler := auth.NewAuthHandler(mockService)

		reqBody := auth.LoginRequest{Email: "test@example.com", Password: "wrongpassword"}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockService.On("Login", mock.Anything, "test@example.com", "wrongpassword").Return(nil, auth.ErrInvalidCredentials)

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var res auth.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, "invalid credentials", res.Message)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockAuthService)
		handler := auth.NewAuthHandler(mockService)

		reqBody := auth.RefreshRequest{RefreshToken: "valid_refresh_token"}
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := &auth.TokenResponse{
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
		}

		mockService.On("Refresh", mock.Anything, "valid_refresh_token").Return(expectedResponse, nil)

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res auth.TokenResponse
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.AccessToken, res.AccessToken)
		assert.Equal(t, expectedResponse.RefreshToken, res.RefreshToken)

		mockService.AssertExpectations(t)
	})

	t.Run("validation_failed", func(t *testing.T) {
		e := setupEcho()
		mockService := new(MockAuthService)
		handler := auth.NewAuthHandler(mockService)

		reqBody := auth.RefreshRequest{} // Missing refresh token
		reqBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
