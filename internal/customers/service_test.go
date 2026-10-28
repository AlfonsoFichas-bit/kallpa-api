package customers

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"kallpa-api/internal/domain"
)

type mockCustomerRepo struct {
	customer *domain.Customer
}

func (m *mockCustomerRepo) FindByTelegramID(ctx context.Context, storeID int64, telegramID int64) (*domain.Customer, error) {
	if m.customer != nil && m.customer.StoreID == storeID && *m.customer.TelegramID == telegramID {
		return m.customer, nil
	}
	return nil, nil
}

func (m *mockCustomerRepo) UpsertByTelegramID(ctx context.Context, storeID int64, telegramID int64, name string) (*domain.Customer, error) {
	if m.customer != nil && m.customer.StoreID == storeID && *m.customer.TelegramID == telegramID {
		m.customer.FullName = &name
		return m.customer, nil
	}

	m.customer = &domain.Customer{
		ID:         1,
		StoreID:    storeID,
		TelegramID: &telegramID,
		FullName:   &name,
		Status:     "lead",
		LTV:        decimal.NewFromInt(0),
		CreatedAt:  time.Now(),
		TelegramIdentity: &domain.TelegramIdentity{
			TelegramID: telegramID,
			FirstName:  &name,
		},
	}
	return m.customer, nil
}

func (m *mockCustomerRepo) ListByStore(ctx context.Context, storeID int64, status string) ([]domain.Customer, error) {
	if m.customer != nil && m.customer.StoreID == storeID {
		if status == "" || m.customer.Status == status {
			return []domain.Customer{*m.customer}, nil
		}
	}
	return []domain.Customer{}, nil
}

func (m *mockCustomerRepo) GetByID(ctx context.Context, storeID int64, customerID int64) (*domain.Customer, error) {
	if m.customer != nil && m.customer.ID == customerID && m.customer.StoreID == storeID {
		return m.customer, nil
	}
	return nil, nil
}

func TestCustomerService_UpsertIdempotency(t *testing.T) {
	repo := &mockCustomerRepo{}
	svc := NewService(repo)
	ctx := context.Background()

	// 1. Create a new customer
	c1, err := svc.UpsertCustomerFromTelegram(ctx, 1, 12345, "John Doe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c1.ID != 1 || *c1.FullName != "John Doe" {
		t.Errorf("unexpected customer values: %+v", c1)
	}

	// 2. Upsert same telegram ID, should return same customer with updated name
	c2, err := svc.UpsertCustomerFromTelegram(ctx, 1, 12345, "John Updated")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c2.ID != c1.ID {
		t.Errorf("expected idempotent customer ID %d, got %d", c1.ID, c2.ID)
	}
	if *c2.FullName != "John Updated" {
		t.Errorf("expected updated name")
	}
}
