package shipping

import (
	"context"

	"kallpa-api/internal/domain"
)

// DeliveryZoneRepository defines operations for the shipping module.
type DeliveryZoneRepository interface {
	ListZones(ctx context.Context, storeID int64) ([]domain.DeliveryZone, error)
	GetByID(ctx context.Context, storeID, zoneID int64) (*domain.DeliveryZone, error)
	CreateZone(ctx context.Context, zone *domain.DeliveryZone) error
	DeleteZone(ctx context.Context, storeID, zoneID int64) error
}
