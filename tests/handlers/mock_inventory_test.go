package handlers_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/domain"
)

type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) SearchProducts(ctx context.Context, storeID int64, query string) ([]domain.Product, error) {
	args := m.Called(ctx, storeID, query)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryService) GetProduct(ctx context.Context, storeID, id int64) (*domain.Product, error) {
	args := m.Called(ctx, storeID, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryService) AdjustStock(ctx context.Context, storeID, variantID int64, delta int) error {
	args := m.Called(ctx, storeID, variantID, delta)
	return args.Error(0)
}

func (m *MockInventoryService) GetLowStock(ctx context.Context, storeID int64, threshold int) ([]domain.ProductVariant, error) {
	args := m.Called(ctx, storeID, threshold)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.ProductVariant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryService) ToggleProductStatus(ctx context.Context, storeID, id int64) error {
	args := m.Called(ctx, storeID, id)
	return args.Error(0)
}
