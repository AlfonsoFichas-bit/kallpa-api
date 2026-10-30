package handlers_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/domain"
)

type MockCustomerService struct {
	mock.Mock
}

func (m *MockCustomerService) UpsertCustomerFromTelegram(ctx context.Context, storeID, telegramID int64, name string) (*domain.Customer, error) {
	args := m.Called(ctx, storeID, telegramID, name)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCustomerService) GetCustomers(ctx context.Context, storeID int64, status string) ([]domain.Customer, error) {
	args := m.Called(ctx, storeID, status)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCustomerService) GetCustomer(ctx context.Context, storeID, id int64) (*domain.Customer, error) {
	args := m.Called(ctx, storeID, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}
