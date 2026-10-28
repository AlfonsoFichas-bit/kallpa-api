package chat

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kallpa-api/internal/domain"
)

// PGRepository implements Repository via GORM.
type PGRepository struct {
	db *gorm.DB
}

// NewRepository creates a new chat repository instance.
func NewRepository(db *gorm.DB) *PGRepository {
	return &PGRepository{db: db}
}

// GetActiveSession retrieves an active chat session for a customer.
func (r *PGRepository) GetActiveSession(ctx context.Context, storeID, customerID int64) (*domain.ChatSession, error) {
	var session domain.ChatSession
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND customer_id = ? AND is_active = ?", storeID, customerID, true).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// UpsertSession creates or updates a chat session.
func (r *PGRepository) UpsertSession(ctx context.Context, session *domain.ChatSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// CreateLog saves an AI log entry.
func (r *PGRepository) CreateLog(ctx context.Context, log *domain.AiLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
