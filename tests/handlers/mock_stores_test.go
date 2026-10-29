package handlers_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/domain"
)

type MockStoresService struct {
	mock.Mock
}

func (m *MockStoresService) GetStoreInfo(ctx context.Context, storeID int64) (*domain.Store, error) {
	args := m.Called(ctx, storeID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Store), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStoresService) ListActiveStores(ctx context.Context) ([]domain.Store, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Store), args.Error(1)
	}
	return nil, args.Error(1)
}
