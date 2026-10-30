package services_test

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/domain"
)

type MockDeliveryZoneRepository struct {
	mock.Mock
}

func (m *MockDeliveryZoneRepository) ListZones(ctx context.Context, storeID int64) ([]domain.DeliveryZone, error) {
	args := m.Called(ctx, storeID)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.DeliveryZone), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDeliveryZoneRepository) GetByID(ctx context.Context, storeID, zoneID int64) (*domain.DeliveryZone, error) {
	args := m.Called(ctx, storeID, zoneID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.DeliveryZone), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDeliveryZoneRepository) CreateZone(ctx context.Context, zone *domain.DeliveryZone) error {
	args := m.Called(ctx, zone)
	if zone != nil && zone.ID == 0 && args.Error(0) == nil {
		zone.ID = 1
	}
	return args.Error(0)
}

func (m *MockDeliveryZoneRepository) DeleteZone(ctx context.Context, storeID, zoneID int64) error {
	args := m.Called(ctx, storeID, zoneID)
	return args.Error(0)
}

type MockSalesService struct {
	mock.Mock
}

func (m *MockSalesService) ListOrders(ctx context.Context, storeID int64, status string) ([]domain.Order, error) {
	args := m.Called(ctx, storeID, status)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSalesService) CreateOrder(ctx context.Context, storeID int64, order *domain.Order) (*domain.Order, error) {
	args := m.Called(ctx, storeID, order)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSalesService) GetOrder(ctx context.Context, storeID, orderID int64) (*domain.Order, error) {
	args := m.Called(ctx, storeID, orderID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSalesService) UpdateOrderStatus(ctx context.Context, storeID, orderID int64, newStatus domain.OrderStatus) error {
	args := m.Called(ctx, storeID, orderID, newStatus)
	return args.Error(0)
}

func (m *MockSalesService) RecalculateOrderTotal(ctx context.Context, storeID, orderID int64, deliveryCost decimal.Decimal) error {
	args := m.Called(ctx, storeID, orderID, deliveryCost)
	return args.Error(0)
}
