package handlers_test

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/domain"
)

type MockShippingService struct {
	mock.Mock
}

func (m *MockShippingService) ListStoreZones(ctx context.Context, storeID int64) ([]domain.DeliveryZone, error) {
	args := m.Called(ctx, storeID)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.DeliveryZone), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockShippingService) AssignZoneToOrder(ctx context.Context, storeID, orderID, zoneID int64) error {
	args := m.Called(ctx, storeID, orderID, zoneID)
	return args.Error(0)
}

func (m *MockShippingService) CreateZone(ctx context.Context, storeID int64, name string, price decimal.Decimal) (*domain.DeliveryZone, error) {
	args := m.Called(ctx, storeID, name, price)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.DeliveryZone), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockShippingService) DeleteZone(ctx context.Context, storeID, zoneID int64) error {
	args := m.Called(ctx, storeID, zoneID)
	return args.Error(0)
}
