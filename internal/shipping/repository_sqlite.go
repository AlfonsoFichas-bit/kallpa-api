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

// CreateZone inserts a new delivery zone.
func (r *DeliveryZonePGRepository) CreateZone(ctx context.Context, zone *domain.DeliveryZone) error {
	return r.db.WithContext(ctx).Create(zone).Error
}

// DeleteZone permanently removes a delivery zone.
func (r *DeliveryZonePGRepository) DeleteZone(ctx context.Context, storeID, zoneID int64) error {
	res := r.db.WithContext(ctx).
		Where("store_id = ? AND id = ?", storeID, zoneID).
		Delete(&domain.DeliveryZone{})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
