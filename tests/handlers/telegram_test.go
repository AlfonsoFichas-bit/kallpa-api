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

	"kallpa-api/internal/telegram"
)

func TestHandleWebhook(t *testing.T) {
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockTelegramService)
		h := telegram.NewHandler(mockSvc)

		update := telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Text:      "Hello",
			},
		}
		jsonBody, _ := json.Marshal(update)

		req := httptest.NewRequest(http.MethodPost, "/webhook/telegram/my-token", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.POST("/webhook/telegram/:token", h.HandleWebhook)
		e.Router().Route(c)

		mockSvc.On("ProcessWebhook", mock.Anything, "my-token", &update).Return(nil)

		err := h.HandleWebhook(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Missing Bot Token", func(t *testing.T) {
		mockSvc := new(MockTelegramService)
		h := telegram.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/webhook/telegram/", bytes.NewReader([]byte("{}")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.POST("/webhook/telegram/:token", h.HandleWebhook)
		e.Router().Route(c)

		err := h.HandleWebhook(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Invalid Payload - Returns 200 OK (Ignored)", func(t *testing.T) {
		mockSvc := new(MockTelegramService)
		h := telegram.NewHandler(mockSvc)

		// Invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/webhook/telegram/my-token", bytes.NewReader([]byte("{invalid json}")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		e.POST("/webhook/telegram/:token", h.HandleWebhook)
		e.Router().Route(c)

		err := h.HandleWebhook(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, "ignored", res["message"])
	})
}
