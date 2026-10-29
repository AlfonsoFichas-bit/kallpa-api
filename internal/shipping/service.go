package shipping

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"

	"kallpa-api/internal/domain"
	"kallpa-api/internal/sales"
)

var (
	// ErrZoneNotFound indicates the delivery zone is not available for this store.
	ErrZoneNotFound = errors.New("delivery zone not found or inactive")
)

// ShippingService defines business operations for shipping and delivery zones.
type ShippingService interface {
	ListStoreZones(ctx context.Context, storeID int64) ([]domain.DeliveryZone, error)
	AssignZoneToOrder(ctx context.Context, storeID, orderID, zoneID int64) error
	CreateZone(ctx context.Context, storeID int64, name string, price decimal.Decimal) (*domain.DeliveryZone, error)
	DeleteZone(ctx context.Context, storeID, zoneID int64) error
}

type service struct {
	repo         DeliveryZoneRepository
	salesService sales.SalesService
}

// NewService creates a new ShippingService.
func NewService(repo DeliveryZoneRepository, salesService sales.SalesService) ShippingService {
	return &service{
		repo:         repo,
		salesService: salesService,
	}
}

// ListStoreZones returns available zones and always includes a virtual "pickup" option.
func (s *service) ListStoreZones(ctx context.Context, storeID int64) ([]domain.DeliveryZone, error) {
	zones, err := s.repo.ListZones(ctx, storeID)
	if err != nil {
		return nil, err
	}

	// Add virtual pickup option
	pickup := domain.DeliveryZone{
		ID:       -1,
		StoreID:  storeID,
		Name:     "Recojo en tienda",
		Price:    decimal.NewFromInt(0),
		IsActive: true,
	}

	return append([]domain.DeliveryZone{pickup}, zones...), nil
}

// AssignZoneToOrder validates the zone and delegates to sales service to recalculate totals.
func (s *service) AssignZoneToOrder(ctx context.Context, storeID, orderID, zoneID int64) error {
	var deliveryCost decimal.Decimal

	if zoneID == -1 {
		// Pickup
		deliveryCost = decimal.NewFromInt(0)
	} else {
		zone, err := s.repo.GetByID(ctx, storeID, zoneID)
		if err != nil {
			return err
		}
		if zone == nil || !zone.IsActive {
			return ErrZoneNotFound
		}
		deliveryCost = zone.Price
	}

	// Tell sales service to update the order
	return s.salesService.RecalculateOrderTotal(ctx, storeID, orderID, deliveryCost)
}

// CreateZone adds a new delivery zone for a store.
func (s *service) CreateZone(ctx context.Context, storeID int64, name string, price decimal.Decimal) (*domain.DeliveryZone, error) {
	zone := &domain.DeliveryZone{
		StoreID:  storeID,
		Name:     name,
		Price:    price,
		IsActive: true,
	}
	if err := s.repo.CreateZone(ctx, zone); err != nil {
		return nil, err
	}
	return zone, nil
}

// DeleteZone removes a delivery zone.
func (s *service) DeleteZone(ctx context.Context, storeID, zoneID int64) error {
	err := s.repo.DeleteZone(ctx, storeID, zoneID)
	if err != nil && err.Error() == "record not found" {
		return ErrZoneNotFound
	}
	return err
}
