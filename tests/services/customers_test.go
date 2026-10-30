package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/customers"
	"kallpa-api/internal/domain"
)

func TestCustomersService_UpsertCustomerFromTelegram(t *testing.T) {
	mockRepo := new(MockCustomerRepository)
	svc := customers.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		name := "Test User"
		expectedCustomer := &domain.Customer{ID: 1, FullName: &name, StoreID: 1}

		mockRepo.On("UpsertByTelegramID", mock.Anything, int64(1), int64(12345), "Test User").Return(expectedCustomer, nil).Once()

		customer, err := svc.UpsertCustomerFromTelegram(context.Background(), 1, 12345, "Test User")

		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, "Test User", *customer.FullName)
	})
}

func TestCustomersService_GetCustomers(t *testing.T) {
	mockRepo := new(MockCustomerRepository)
	svc := customers.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		name := "Test User"
		expectedCustomers := []domain.Customer{
			{ID: 1, FullName: &name},
		}

		mockRepo.On("ListByStore", mock.Anything, int64(1), "active").Return(expectedCustomers, nil).Once()

		list, err := svc.GetCustomers(context.Background(), 1, "active")

		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "Test User", *list[0].FullName)
	})
}

func TestCustomersService_GetCustomer(t *testing.T) {
	mockRepo := new(MockCustomerRepository)
	svc := customers.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		name := "Test User"
		expectedCustomer := &domain.Customer{ID: 1, FullName: &name}

		mockRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(expectedCustomer, nil).Once()

		customer, err := svc.GetCustomer(context.Background(), 1, 1)

		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, int64(1), customer.ID)
	})
}
