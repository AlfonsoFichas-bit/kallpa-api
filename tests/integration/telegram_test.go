package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	"kallpa-api/internal/domain"
	"kallpa-api/internal/telegram"
)

func TestIntegration_TelegramWebhook(t *testing.T) {
	// Seed a store with a mock bot token
	testDB.Exec("DELETE FROM stores")
	botToken := "test-bot-token"
	store := &domain.Store{
		ID:        1,
		Name:      "Test Store",
		Slug:      "test-store",
		BotToken:  &botToken,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	testDB.Create(store)

	// Build a dummy telegram update
	update := telegram.Update{
		UpdateID: 100,
		Message: &telegram.Message{
			MessageID: 1,
			From: &telegram.User{
				ID:        999,
				FirstName: "Jane",
				Username:  "jane_doe",
			},
			Chat: &telegram.Chat{
				ID: 999,
			},
			Text: "/start",
		},
	}

	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPost, "/v1/webhook/telegram/"+botToken, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	testRouter.ServeHTTP(rec, req)

	// Since we mock the client implicitly by letting it hit a dummy endpoint or fail gracefully,
	// we just want to ensure our webhook processes without crashing and returns 200 to Telegram.
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for telegram webhook, got %d", rec.Code)
	}
}
