package notifications

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"kallpa-api/internal/domain"
)

// PGRepository implements Repository via Gorm.
type PGRepository struct {
	db *gorm.DB
}

// NewRepository initializes a new notification repository.
func NewRepository(db *gorm.DB) *PGRepository {
	return &PGRepository{db: db}
}

// ListNotifications retrieves notifications for a store.
func (r *PGRepository) ListNotifications(ctx context.Context, storeID int64, unreadOnly bool) ([]domain.Notification, error) {
	var notifications []domain.Notification
	q := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if unreadOnly {
		q = q.Where("is_read = ?", false)
	}
	err := q.Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

// MarkAsRead sets a notification's is_read flag to true.
func (r *PGRepository) MarkAsRead(ctx context.Context, storeID, notificationID int64) error {
	res := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("id = ? AND store_id = ?", notificationID, storeID).
		Update("is_read", true)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("notification not found")
	}
	return nil
}

// Create inserts a new notification into the database.
func (r *PGRepository) Create(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}
