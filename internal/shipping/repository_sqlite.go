package shipping

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"kallpa-api/internal/domain"
)

// DeliveryZonePGRepository handles database operations for delivery zones.
type DeliveryZonePGRepository struct {
	db *gorm.DB
}

// NewDeliveryZoneRepository creates a new instance.
func NewDeliveryZoneRepository(db *gorm.DB) *DeliveryZonePGRepository {
	return &DeliveryZonePGRepository{db: db}
}

// ListZones returns all active delivery zones for a store.
func (r *DeliveryZonePGRepository) ListZones(ctx context.Context, storeID int64) ([]domain.DeliveryZone, error) {
	var zones []domain.DeliveryZone
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND is_active = ?", storeID, true).
		Find(&zones).Error
	return zones, err
}

// GetByID returns a specific delivery zone.
func (r *DeliveryZonePGRepository) GetByID(ctx context.Context, storeID, zoneID int64) (*domain.DeliveryZone, error) {
	var zone domain.DeliveryZone
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND id = ?", storeID, zoneID).
		First(&zone).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &zone, nil
}
