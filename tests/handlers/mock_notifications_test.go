package handlers_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/domain"
)

type MockNotificationsService struct {
	mock.Mock
}

func (m *MockNotificationsService) GetNotifications(ctx context.Context, storeID int64, unreadOnly bool) ([]domain.Notification, error) {
	args := m.Called(ctx, storeID, unreadOnly)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Notification), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNotificationsService) MarkNotificationAsRead(ctx context.Context, storeID, notificationID int64) error {
	args := m.Called(ctx, storeID, notificationID)
	return args.Error(0)
}

func (m *MockNotificationsService) EmitNotification(ctx context.Context, storeID int64, nType domain.NotificationType, msg string, orderID *int64) error {
	args := m.Called(ctx, storeID, nType, msg, orderID)
	return args.Error(0)
}
