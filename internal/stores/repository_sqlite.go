package stores

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

// NewRepository initializes a new store repository.
func NewRepository(db *gorm.DB) *PGRepository {
	return &PGRepository{db: db}
}

// GetByID retrieves a specific store by ID.
func (r *PGRepository) GetByID(ctx context.Context, storeID int64) (*domain.Store, error) {
	var store domain.Store
	err := r.db.WithContext(ctx).First(&store, storeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &store, nil
}

// ListAll retrieves all active stores (primarily for superadmin dashboard).
func (r *PGRepository) ListAll(ctx context.Context) ([]domain.Store, error) {
	var stores []domain.Store
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&stores).Error
	return stores, err
}
