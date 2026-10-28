package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"kallpa-api/internal/chat"
	"kallpa-api/internal/domain"
)

func TestIntegration_ChatFlow(t *testing.T) {
	_, _ = setupStoreData(t) // Initializes store and users
	token := getAdminAuthToken(t)

	// Clean up chat data before testing
	testDB.Exec("DELETE FROM ai_logs")
	testDB.Exec("DELETE FROM chat_sessions")

	// 1. Process Message
	chatReq := chat.MessageRequest{
		StoreID:        1,
		ChatID:         "chat-123",
		TelegramUserID: 112233,
		UserName:       "Test Cust",
		Message:        "Quiero 2 bolsas de quinua",
	}

	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/conversations/message", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for processing message, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var res chat.MessageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify mocked AI response
	if res.Intent != "add_to_cart" {
		t.Errorf("expected intent 'add_to_cart', got '%s'", res.Intent)
	}
	if !res.RequiresConfirmation {
		t.Errorf("expected requires_confirmation to be true")
	}

	// 2. Verify Session Creation
	var count int64
	testDB.Model(&domain.ChatSession{}).Where("store_id = ?", 1).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 chat session, got %d", count)
	}

	var session domain.ChatSession
	testDB.First(&session)
	if *session.CurrentStep != "add_to_cart" {
		t.Errorf("expected current step to be 'add_to_cart', got %v", *session.CurrentStep)
	}

	// 3. Verify AI Log
	testDB.Model(&domain.AiLog{}).Where("store_id = ?", 1).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 ai log, got %d", count)
	}

	var aiLog domain.AiLog
	testDB.First(&aiLog)
	if *aiLog.ModelUsed != "dummy-v1" {
		t.Errorf("expected model used 'dummy-v1', got %v", *aiLog.ModelUsed)
	}
	if *aiLog.ChatSessionID != session.ID {
		t.Errorf("expected ai log session ID %d, got %v", session.ID, *aiLog.ChatSessionID)
	}

	// 4. Process Second Message (Existing Session)
	chatReq2 := chat.MessageRequest{
		StoreID:        1,
		ChatID:         "chat-123",
		TelegramUserID: 112233, // Same user
		UserName:       "Test Cust",
		Message:        "Si", // Triggers 'confirm_order'
	}
	body2, _ := json.Marshal(chatReq2)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/conversations/message", bytes.NewReader(body2))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()

	testRouter.ServeHTTP(rec2, req2)

	// Verify only 1 session still exists (Idempotent)
	testDB.Model(&domain.ChatSession{}).Where("store_id = ?", 1).Count(&count)
	if count != 1 {
		t.Errorf("expected still 1 chat session, got %d", count)
	}

	// Verify a 2nd AI Log was created
	testDB.Model(&domain.AiLog{}).Where("store_id = ?", 1).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 ai logs, got %d", count)
	}
}
