package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"kallpa-api/internal/domain"
)

func TestIntegration_NotificationAndStoreFlow(t *testing.T) {
	_, _ = setupStoreData(t)
	token := getAdminAuthToken(t)

	// Clean up tables
	testDB.Exec("DELETE FROM notifications")
	testDB.Exec("DELETE FROM stores")

	// Seed a store
	store := &domain.Store{
		ID:        1,
		Name:      "Test Store",
		Slug:      "test-store",
		Currency:  "BOB",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := testDB.Create(store).Error; err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}

	// Seed a notification
	notif := &domain.Notification{
		StoreID:   1,
		Type:      domain.NotificationTypeSale,
		Message:   "New sale completed",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	if err := testDB.Create(notif).Error; err != nil {
		t.Fatalf("failed to seed notification: %v", err)
	}

	// 1. Get Store Info
	req := httptest.NewRequest(http.MethodGet, "/v1/stores/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for get store, got %d. Body: %s", rec.Code, rec.Body.String())
	}
	var fetchedStore domain.Store
	json.Unmarshal(rec.Body.Bytes(), &fetchedStore)
	if fetchedStore.Name != "Test Store" {
		t.Errorf("expected store name 'Test Store', got '%s'", fetchedStore.Name)
	}

	// 2. List Notifications (Unread)
	req = httptest.NewRequest(http.MethodGet, "/v1/notifications?unread=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for list notifications, got %d", rec.Code)
	}
	var notifs []domain.Notification
	json.Unmarshal(rec.Body.Bytes(), &notifs)
	if len(notifs) != 1 {
		t.Errorf("expected 1 unread notification, got %d", len(notifs))
	}
	if notifs[0].Message != "New sale completed" {
		t.Errorf("expected message 'New sale completed', got '%s'", notifs[0].Message)
	}

	// 3. Mark Notification as Read
	req = httptest.NewRequest(http.MethodPatch, "/v1/notifications/"+strconv.FormatInt(notif.ID, 10)+"/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for mark as read, got %d", rec.Code)
	}

	// 4. Verify Notification is Read
	var count int64
	testDB.Model(&domain.Notification{}).Where("is_read = ?", true).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 read notification in DB, got %d", count)
	}
}
