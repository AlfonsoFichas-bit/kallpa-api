package notifications

import (
	"context"

	"kallpa-api/internal/domain"
)

// Service defines business operations for notifications.
type Service interface {
	GetNotifications(ctx context.Context, storeID int64, unreadOnly bool) ([]domain.Notification, error)
	MarkNotificationAsRead(ctx context.Context, storeID, notificationID int64) error
	// EmitNotification is internal business logic used by other services (e.g. Sales, Chat)
	EmitNotification(ctx context.Context, storeID int64, nType domain.NotificationType, msg string, orderID *int64) error
}

type service struct {
	repo Repository
}

// NewService creates a new notification service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// GetNotifications returns notifications for the given store.
func (s *service) GetNotifications(ctx context.Context, storeID int64, unreadOnly bool) ([]domain.Notification, error) {
	return s.repo.ListNotifications(ctx, storeID, unreadOnly)
}

// MarkNotificationAsRead updates the notification status.
func (s *service) MarkNotificationAsRead(ctx context.Context, storeID, notificationID int64) error {
	return s.repo.MarkAsRead(ctx, storeID, notificationID)
}

// EmitNotification creates a new system notification.
func (s *service) EmitNotification(ctx context.Context, storeID int64, nType domain.NotificationType, msg string, orderID *int64) error {
	n := &domain.Notification{
		StoreID: storeID,
		Type:    nType,
		Message: msg,
		OrderID: orderID,
	}
	return s.repo.Create(ctx, n)
}
