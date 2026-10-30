package services_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/domain"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) ListOrders(ctx context.Context, storeID int64, status string) ([]domain.Order, error) {
	args := m.Called(ctx, storeID, status)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	// Optionally set an ID on creation to simulate DB behavior if passed a pointer
	if order != nil && order.ID == 0 && args.Error(0) == nil {
		order.ID = 1
	}
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, storeID, orderID int64) (*domain.Order, error) {
	args := m.Called(ctx, storeID, orderID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, storeID, orderID int64, newStatus domain.OrderStatus) error {
	args := m.Called(ctx, storeID, orderID, newStatus)
	return args.Error(0)
}

func (m *MockOrderRepository) FindLatestDraftByCustomer(ctx context.Context, storeID, customerID int64) (*domain.Order, error) {
	args := m.Called(ctx, storeID, customerID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) FindByIdempotencyKey(ctx context.Context, storeID int64, key string) (*domain.Order, error) {
	args := m.Called(ctx, storeID, key)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) FindByTelegramID(ctx context.Context, storeID, telegramID int64) (*domain.Customer, error) {
	args := m.Called(ctx, storeID, telegramID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCustomerRepository) UpsertByTelegramID(ctx context.Context, storeID, telegramID int64, name string) (*domain.Customer, error) {
	args := m.Called(ctx, storeID, telegramID, name)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCustomerRepository) ListByStore(ctx context.Context, storeID int64, status string) ([]domain.Customer, error) {
	args := m.Called(ctx, storeID, status)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCustomerRepository) GetByID(ctx context.Context, storeID, customerID int64) (*domain.Customer, error) {
	args := m.Called(ctx, storeID, customerID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}
