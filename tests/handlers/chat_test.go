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
	"kallpa-api/internal/chat"
)

// ChatCustomValidator for echo
type ChatCustomValidator struct {
	validator *validator.Validate
}

func (cv *ChatCustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func TestProcessMessage(t *testing.T) {
	e := echo.New()
	e.Validator = &ChatCustomValidator{validator: validator.New()}

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockChatService)
		h := chat.NewHandler(mockSvc)

		reqBody := chat.MessageRequest{
			StoreID:        1,
			ChatID:         "12345",
			TelegramUserID: 98765,
			UserName:       "TestUser",
			Message:        "Hello",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/conversations/message", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		expectedRes := &chat.MessageResponse{
			Reply:  "Hi there!",
			Intent: "greeting",
		}

		// Expected request will have StoreID overwritten by token StoreID
		expectedReq := reqBody
		expectedReq.StoreID = storeID

		mockSvc.On("ProcessMessage", mock.Anything, &expectedReq).Return(expectedRes, nil)

		err := h.ProcessMessage(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res chat.MessageResponse
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, expectedRes.Reply, res.Reply)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Unauthorized - No Claims", func(t *testing.T) {
		mockSvc := new(MockChatService)
		h := chat.NewHandler(mockSvc)

		reqBody := chat.MessageRequest{
			StoreID:        1,
			ChatID:         "12345",
			TelegramUserID: 98765,
			UserName:       "TestUser",
			Message:        "Hello",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/conversations/message", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.ProcessMessage(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Unauthorized - No StoreID", func(t *testing.T) {
		mockSvc := new(MockChatService)
		h := chat.NewHandler(mockSvc)

		reqBody := chat.MessageRequest{
			StoreID:        1,
			ChatID:         "12345",
			TelegramUserID: 98765,
			UserName:       "TestUser",
			Message:        "Hello",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/conversations/message", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		claims := &auth.CustomClaims{StoreID: nil}
		c.Set(auth.ContextKeyUser, claims)

		err := h.ProcessMessage(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		mockSvc := new(MockChatService)
		h := chat.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/conversations/message", bytes.NewReader([]byte("{invalid json}")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.ProcessMessage(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Validation Failed - Missing Required Fields", func(t *testing.T) {
		mockSvc := new(MockChatService)
		h := chat.NewHandler(mockSvc)

		reqBody := chat.MessageRequest{
			// Missing ChatID, TelegramUserID, UserName, Message
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/conversations/message", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.ProcessMessage(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockChatService)
		h := chat.NewHandler(mockSvc)

		reqBody := chat.MessageRequest{
			StoreID:        1,
			ChatID:         "12345",
			TelegramUserID: 98765,
			UserName:       "TestUser",
			Message:        "Hello",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/conversations/message", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		expectedReq := reqBody
		expectedReq.StoreID = storeID

		mockSvc.On("ProcessMessage", mock.Anything, &expectedReq).Return(nil, errors.New("ai error"))

		err := h.ProcessMessage(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		mockSvc.AssertExpectations(t)
	})
}
