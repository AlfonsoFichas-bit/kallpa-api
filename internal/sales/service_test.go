package sales

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
)

type mockOrderRepo struct {
	orders []domain.Order
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, order *domain.Order) error {
	order.ID = int64(len(m.orders) + 1)
	m.orders = append(m.orders, *order)
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, storeID, orderID int64) (*domain.Order, error) {
	for i := range m.orders {
		if m.orders[i].ID == orderID && m.orders[i].StoreID == storeID {
			return &m.orders[i], nil
		}
	}
	return nil, nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, storeID, orderID int64, newStatus domain.OrderStatus) error {
	for i := range m.orders {
		if m.orders[i].ID == orderID && m.orders[i].StoreID == storeID {
			m.orders[i].Status = newStatus
			return nil
		}
	}
	return nil
}

func (m *mockOrderRepo) Update(ctx context.Context, order *domain.Order) error {
	for i := range m.orders {
		if m.orders[i].ID == order.ID && m.orders[i].StoreID == order.StoreID {
			m.orders[i] = *order
			return nil
		}
	}
	return nil
}

func (m *mockOrderRepo) FindLatestDraftByCustomer(ctx context.Context, storeID, customerID int64) (*domain.Order, error) {
	return nil, nil
}

func (m *mockOrderRepo) FindByIdempotencyKey(ctx context.Context, storeID int64, key string) (*domain.Order, error) {
	for i := range m.orders {
		if m.orders[i].IdempotencyKey != nil && *m.orders[i].IdempotencyKey == key && m.orders[i].StoreID == storeID {
			return &m.orders[i], nil
		}
	}
	return nil, nil
}

type mockProductRepo struct {
}

func (m *mockProductRepo) ListProducts(ctx context.Context, storeID int64, query inventory.ProductSearchQuery) ([]domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) GetProductByID(ctx context.Context, storeID, productID int64) (*domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) GetVariantByID(ctx context.Context, variantID int64) (*domain.ProductVariant, error) {
	prod := &domain.Product{
		BasePrice: decimal.NewFromInt(10),
	}
	// Let's vary the additional price based on variant ID to simulate different prices
	additional := decimal.NewFromInt(0)
	if variantID == 2 {
		additional = decimal.NewFromInt(5)
	}

	return &domain.ProductVariant{
		ID:              variantID,
		ProductID:       1,
		Product:         prod,
		AdditionalPrice: additional,
	}, nil
}
func (m *mockProductRepo) UpdateStock(ctx context.Context, variantID int64, delta int) error {
	return nil
}
func (m *mockProductRepo) LowStockQuery(ctx context.Context, storeID int64, threshold int) ([]domain.ProductVariant, error) {
	return nil, nil
}

func TestSalesService_CreateOrder_Totals(t *testing.T) {
	repo := &mockOrderRepo{}
	svc := NewService(repo, &mockProductRepo{})
	ctx := context.Background()

	vid1 := int64(1)
	vid2 := int64(2)
	order := &domain.Order{
		Items: []domain.OrderItem{
			{ProductVariantID: &vid1, Quantity: 2}, // Price will be 10 + 0 = 10
			{ProductVariantID: &vid2, Quantity: 1}, // Price will be 10 + 5 = 15
		},
		DeliveryCost: decimal.NewFromInt(10),
	}

	created, err := svc.CreateOrder(ctx, 1, order)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedSubtotal := decimal.NewFromInt(35) // (2*10) + (1*15) = 20 + 15 = 35
	expectedTotal := decimal.NewFromInt(45)    // 35 + 10 = 45

	if !created.Subtotal.Equal(expectedSubtotal) {
		t.Errorf("expected subtotal %s, got %s", expectedSubtotal, created.Subtotal)
	}
	if !created.Total.Equal(expectedTotal) {
		t.Errorf("expected total %s, got %s", expectedTotal, created.Total)
	}
}

func TestSalesService_ValidStatusTransitions(t *testing.T) {
	repo := &mockOrderRepo{
		orders: []domain.Order{
			{ID: 1, StoreID: 1, Status: domain.OrderStatusDraft},
			{ID: 2, StoreID: 1, Status: domain.OrderStatusDelivered},
		},
	}
	svc := NewService(repo, &mockProductRepo{})
	ctx := context.Background()

	// Valid transition: draft -> pending_payment
	err := svc.UpdateOrderStatus(ctx, 1, 1, domain.OrderStatusPendingPayment)
	if err != nil {
		t.Errorf("expected no error for valid transition, got %v", err)
	}

	// Invalid transition: delivered -> cancelled
	err = svc.UpdateOrderStatus(ctx, 1, 2, domain.OrderStatusCancelled)
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

func TestSalesService_Idempotency(t *testing.T) {
	idempotencyKey := "key-123"
	repo := &mockOrderRepo{
		orders: []domain.Order{
			{ID: 1, StoreID: 1, IdempotencyKey: &idempotencyKey},
		},
	}
	svc := NewService(repo, &mockProductRepo{})
	ctx := context.Background()

	order := &domain.Order{
		IdempotencyKey: &idempotencyKey,
	}
	created, err := svc.CreateOrder(ctx, 1, order)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.ID != 1 {
		t.Errorf("expected existing order with ID 1, got %d", created.ID)
	}
}
