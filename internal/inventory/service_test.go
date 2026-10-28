package inventory

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"kallpa-api/internal/domain"
)

type mockProductRepo struct {
	products []domain.Product
	variants []domain.ProductVariant
}

func (m *mockProductRepo) ListProducts(ctx context.Context, storeID int64, query ProductSearchQuery) ([]domain.Product, error) {
	var result []domain.Product
	for i := range m.products {
		if m.products[i].StoreID == storeID {
			result = append(result, m.products[i]) // For tests, ignore query logic matching to keep it simple
		}
	}
	return result, nil
}

func (m *mockProductRepo) GetProductByID(ctx context.Context, storeID, productID int64) (*domain.Product, error) {
	for i := range m.products {
		if m.products[i].ID == productID && m.products[i].StoreID == storeID {
			return &m.products[i], nil
		}
	}
	return nil, nil
}

func (m *mockProductRepo) GetVariantByID(ctx context.Context, variantID int64) (*domain.ProductVariant, error) {
	for i := range m.variants {
		if m.variants[i].ID == variantID {
			return &m.variants[i], nil
		}
	}
	return nil, nil
}

func (m *mockProductRepo) UpdateStock(ctx context.Context, variantID int64, delta int) error {
	for i := range m.variants {
		if m.variants[i].ID == variantID {
			if m.variants[i].StockQuantity+delta < 0 {
				return ErrInsufficientStock
			}
			m.variants[i].StockQuantity += delta
			return nil
		}
	}
	return nil
}

func (m *mockProductRepo) LowStockQuery(ctx context.Context, storeID int64, threshold int) ([]domain.ProductVariant, error) {
	return []domain.ProductVariant{}, nil
}

func TestInventoryService_AdjustStock(t *testing.T) {
	repo := &mockProductRepo{
		products: []domain.Product{
			{ID: 10, StoreID: 1, Name: "Quinoa", BasePrice: decimal.NewFromInt(50)},
		},
		variants: []domain.ProductVariant{
			{ID: 100, ProductID: 10, StockQuantity: 5},
		},
	}
	svc := NewService(repo)
	ctx := context.Background()

	// 1. Success decrease
	err := svc.AdjustStock(ctx, 1, 100, -3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.variants[0].StockQuantity != 2 {
		t.Errorf("expected stock to be 2, got %d", repo.variants[0].StockQuantity)
	}

	// 2. Fail decrease (insufficient)
	err = svc.AdjustStock(ctx, 1, 100, -3)
	if !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}

	// 3. Success increase
	err = svc.AdjustStock(ctx, 1, 100, 10)
	if err != nil {
		t.Fatalf("expected no error on increase, got %v", err)
	}
	if repo.variants[0].StockQuantity != 12 {
		t.Errorf("expected stock to be 12, got %d", repo.variants[0].StockQuantity)
	}

	// 4. Variant belonging to different store (store mismatch)
	err = svc.AdjustStock(ctx, 999, 100, -1)
	if err == nil || err.Error() != "variant does not belong to this store" {
		t.Errorf("expected store mismatch error, got %v", err)
	}
}
