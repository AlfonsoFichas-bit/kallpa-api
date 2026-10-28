package notifications

import (
	"context"

	"kallpa-api/internal/domain"
)

// Repository defines data operations for notifications.
type Repository interface {
	ListNotifications(ctx context.Context, storeID int64, unreadOnly bool) ([]domain.Notification, error)
	MarkAsRead(ctx context.Context, storeID, notificationID int64) error
	Create(ctx context.Context, notification *domain.Notification) error
}
