package shipping

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/sales"
)

type mockDeliveryZoneRepo struct {
	zones []domain.DeliveryZone
}

func (m *mockDeliveryZoneRepo) ListZones(ctx context.Context, storeID int64) ([]domain.DeliveryZone, error) {
	return m.zones, nil
}

func (m *mockDeliveryZoneRepo) GetByID(ctx context.Context, storeID, zoneID int64) (*domain.DeliveryZone, error) {
	for i := range m.zones {
		if m.zones[i].ID == zoneID && m.zones[i].StoreID == storeID {
			return &m.zones[i], nil
		}
	}
	return nil, nil
}

// Mock Sales Service just for testing RecalculateOrderTotal
type mockSalesService struct {
	sales.SalesService // Embed interface to satisfy it, only override what we need
	calledOrder        int64
	calledDeliveryCost decimal.Decimal
}

func (m *mockSalesService) RecalculateOrderTotal(ctx context.Context, storeID, orderID int64, deliveryCost decimal.Decimal) error {
	m.calledOrder = orderID
	m.calledDeliveryCost = deliveryCost
	return nil
}

func TestShippingService_ListStoreZones(t *testing.T) {
	repo := &mockDeliveryZoneRepo{
		zones: []domain.DeliveryZone{
			{ID: 1, StoreID: 1, Name: "Zone 1", IsActive: true},
		},
	}
	svc := NewService(repo, nil)

	zones, err := svc.ListStoreZones(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(zones) != 2 {
		t.Fatalf("expected 2 zones (1 virtual + 1 real), got %d", len(zones))
	}
	if zones[0].ID != -1 {
		t.Errorf("first zone should be virtual pickup")
	}
}

func TestShippingService_AssignZoneToOrder(t *testing.T) {
	repo := &mockDeliveryZoneRepo{
		zones: []domain.DeliveryZone{
			{ID: 1, StoreID: 1, Name: "Zone 1", Price: decimal.NewFromInt(15), IsActive: true},
		},
	}
	salesSvc := &mockSalesService{}
	svc := NewService(repo, salesSvc)

	// Valid zone assignment
	err := svc.AssignZoneToOrder(context.Background(), 1, 100, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if salesSvc.calledOrder != 100 {
		t.Errorf("expected order ID 100 to be called, got %d", salesSvc.calledOrder)
	}
	if !salesSvc.calledDeliveryCost.Equal(decimal.NewFromInt(15)) {
		t.Errorf("expected delivery cost 15, got %s", salesSvc.calledDeliveryCost)
	}

	// Pickup assignment
	err = svc.AssignZoneToOrder(context.Background(), 1, 101, -1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if salesSvc.calledOrder != 101 {
		t.Errorf("expected order ID 101 to be called, got %d", salesSvc.calledOrder)
	}
	if !salesSvc.calledDeliveryCost.Equal(decimal.NewFromInt(0)) {
		t.Errorf("expected delivery cost 0, got %s", salesSvc.calledDeliveryCost)
	}

	// Invalid zone assignment
	err = svc.AssignZoneToOrder(context.Background(), 1, 102, 99)
	if !errors.Is(err, ErrZoneNotFound) {
		t.Errorf("expected ErrZoneNotFound, got %v", err)
	}
}
