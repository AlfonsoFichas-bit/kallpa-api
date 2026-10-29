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
	"kallpa-api/internal/notifications"
)

func TestListNotifications(t *testing.T) {
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockNotificationsService)
		h := notifications.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/notifications?unread=true", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		expectedNotifs := []domain.Notification{
			{ID: 1, StoreID: storeID, Message: "Test Notif", IsRead: false},
		}

		mockSvc.On("GetNotifications", mock.Anything, storeID, true).Return(expectedNotifs, nil)

		err := h.ListNotifications(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res []domain.Notification
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, expectedNotifs[0].Message, res[0].Message)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Unauthorized - No Claims", func(t *testing.T) {
		mockSvc := new(MockNotificationsService)
		h := notifications.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Don't set user claim

		err := h.ListNotifications(c)
		assert.NoError(t, err) // Echo handles the JSON response, handler itself shouldn't return error here
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Unauthorized - No StoreID", func(t *testing.T) {
		mockSvc := new(MockNotificationsService)
		h := notifications.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		claims := &auth.CustomClaims{StoreID: nil}
		c.Set(auth.ContextKeyUser, claims)

		err := h.ListNotifications(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockNotificationsService)
		h := notifications.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("GetNotifications", mock.Anything, storeID, false).Return(nil, errors.New("db error"))

		err := h.ListNotifications(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		mockSvc.AssertExpectations(t)
	})
}

func TestMarkAsRead(t *testing.T) {
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockNotificationsService)
		h := notifications.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPatch, "/notifications/1/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.PATCH("/notifications/:id/read", h.MarkAsRead)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("MarkNotificationAsRead", mock.Anything, storeID, int64(1)).Return(nil)

		err := h.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Unauthorized - No Claims", func(t *testing.T) {
		mockSvc := new(MockNotificationsService)
		h := notifications.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPatch, "/notifications/1/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.PATCH("/notifications/:id/read", h.MarkAsRead)
		e.Router().Route(c)

		err := h.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		mockSvc := new(MockNotificationsService)
		h := notifications.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPatch, "/notifications/abc/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.PATCH("/notifications/:id/read", h.MarkAsRead)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		err := h.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockSvc := new(MockNotificationsService)
		h := notifications.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPatch, "/notifications/999/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.PATCH("/notifications/:id/read", h.MarkAsRead)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("MarkNotificationAsRead", mock.Anything, storeID, int64(999)).Return(errors.New("notification not found"))

		err := h.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc := new(MockNotificationsService)
		h := notifications.NewHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPatch, "/notifications/1/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.PATCH("/notifications/:id/read", h.MarkAsRead)
		e.Router().Route(c)

		storeID := int64(1)
		claims := &auth.CustomClaims{StoreID: &storeID}
		c.Set(auth.ContextKeyUser, claims)

		mockSvc.On("MarkNotificationAsRead", mock.Anything, storeID, int64(1)).Return(errors.New("db error"))

		err := h.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		mockSvc.AssertExpectations(t)
	})
}
