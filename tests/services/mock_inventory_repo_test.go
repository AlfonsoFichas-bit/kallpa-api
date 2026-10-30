package services_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
)

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) ListProducts(ctx context.Context, storeID int64, query inventory.ProductSearchQuery) ([]domain.Product, error) {
	args := m.Called(ctx, storeID, query)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductRepository) GetProductByID(ctx context.Context, storeID, productID int64) (*domain.Product, error) {
	args := m.Called(ctx, storeID, productID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductRepository) GetVariantByID(ctx context.Context, variantID int64) (*domain.ProductVariant, error) {
	args := m.Called(ctx, variantID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.ProductVariant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductRepository) UpdateStock(ctx context.Context, variantID int64, delta int) error {
	args := m.Called(ctx, variantID, delta)
	return args.Error(0)
}

func (m *MockProductRepository) LowStockQuery(ctx context.Context, storeID int64, threshold int) ([]domain.ProductVariant, error) {
	args := m.Called(ctx, storeID, threshold)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.ProductVariant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductRepository) ToggleProductStatus(ctx context.Context, storeID, productID int64) error {
	args := m.Called(ctx, storeID, productID)
	return args.Error(0)
}
