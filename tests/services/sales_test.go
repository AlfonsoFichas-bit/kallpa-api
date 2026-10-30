package services_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/domain"
	"kallpa-api/internal/sales"
)

func TestSalesService_ListOrders(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockInvRepo := new(MockProductRepository)
	mockCustRepo := new(MockCustomerRepository)
	svc := sales.NewService(mockRepo, mockInvRepo, mockCustRepo)

	t.Run("Success", func(t *testing.T) {
		expectedOrders := []domain.Order{
			{ID: 1, StoreID: 1, Status: domain.OrderStatusPaid},
		}

		mockRepo.On("ListOrders", mock.Anything, int64(1), string(domain.OrderStatusPaid)).Return(expectedOrders, nil).Once()

		orders, err := svc.ListOrders(context.Background(), 1, string(domain.OrderStatusPaid))
		assert.NoError(t, err)
		assert.Len(t, orders, 1)
	})
}

func TestSalesService_CreateOrder(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockInvRepo := new(MockProductRepository)
	mockCustRepo := new(MockCustomerRepository)
	svc := sales.NewService(mockRepo, mockInvRepo, mockCustRepo)

	t.Run("Success", func(t *testing.T) {
		variantID := int64(10)
		order := &domain.Order{
			CustomerID: 20,
			Items: []domain.OrderItem{
				{ProductVariantID: &variantID, Quantity: 2},
			},
		}

		// Mock Customer
		mockCustRepo.On("GetByID", mock.Anything, int64(1), int64(20)).Return(&domain.Customer{ID: 20, StoreID: 1}, nil).Once()

		// Mock Variant
		variant := &domain.ProductVariant{
			ID: 10,
			AdditionalPrice: decimal.NewFromFloat(5.0),
			Product: &domain.Product{
				ID: 100, StoreID: 1, BasePrice: decimal.NewFromFloat(10.0), Name: "Test Product",
			},
		}
		mockInvRepo.On("GetVariantByID", mock.Anything, int64(10)).Return(variant, nil).Once()

		// Mock Create
		mockRepo.On("CreateOrder", mock.Anything, mock.MatchedBy(func(o *domain.Order) bool {
			return o.Subtotal.Equal(decimal.NewFromFloat(30.0)) // 2 * (10 + 5)
		})).Return(nil).Once()

		createdOrder, err := svc.CreateOrder(context.Background(), 1, order)

		assert.NoError(t, err)
		assert.NotNil(t, createdOrder)
		assert.True(t, createdOrder.Subtotal.Equal(decimal.NewFromFloat(30.0)))
	})

	t.Run("Idempotency Success", func(t *testing.T) {
		key := "unique-key"
		order := &domain.Order{
			CustomerID: 20,
			IdempotencyKey: &key,
		}

		mockCustRepo.On("GetByID", mock.Anything, int64(1), int64(20)).Return(&domain.Customer{ID: 20, StoreID: 1}, nil).Once()

		existingOrder := &domain.Order{ID: 99, IdempotencyKey: &key}
		mockRepo.On("FindByIdempotencyKey", mock.Anything, int64(1), key).Return(existingOrder, nil).Once()

		createdOrder, err := svc.CreateOrder(context.Background(), 1, order)

		assert.NoError(t, err)
		assert.Equal(t, int64(99), createdOrder.ID)
		// It should bypass creating the order again
	})

	t.Run("Customer Not Found", func(t *testing.T) {
		order := &domain.Order{CustomerID: 20}
		mockCustRepo.On("GetByID", mock.Anything, int64(1), int64(20)).Return(nil, nil).Once()

		createdOrder, err := svc.CreateOrder(context.Background(), 1, order)

		assert.ErrorContains(t, err, "customer 20 not found")
		assert.Nil(t, createdOrder)
	})

	t.Run("Variant Belongs to Another Store", func(t *testing.T) {
		variantID := int64(10)
		order := &domain.Order{
			CustomerID: 20,
			Items: []domain.OrderItem{
				{ProductVariantID: &variantID, Quantity: 2},
			},
		}

		mockCustRepo.On("GetByID", mock.Anything, int64(1), int64(20)).Return(&domain.Customer{ID: 20, StoreID: 1}, nil).Once()

		variant := &domain.ProductVariant{
			ID: 10,
			Product: &domain.Product{
				ID: 100, StoreID: 99, // Wrong store ID
			},
		}
		mockInvRepo.On("GetVariantByID", mock.Anything, int64(10)).Return(variant, nil).Once()

		createdOrder, err := svc.CreateOrder(context.Background(), 1, order)

		assert.ErrorContains(t, err, "does not belong to this store")
		assert.Nil(t, createdOrder)
	})
}

func TestSalesService_GetOrder(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockInvRepo := new(MockProductRepository)
	mockCustRepo := new(MockCustomerRepository)
	svc := sales.NewService(mockRepo, mockInvRepo, mockCustRepo)

	t.Run("Success", func(t *testing.T) {
		expectedOrder := &domain.Order{ID: 1, StoreID: 1}
		mockRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(expectedOrder, nil).Once()

		order, err := svc.GetOrder(context.Background(), 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), order.ID)
	})
}

func TestSalesService_UpdateOrderStatus(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockInvRepo := new(MockProductRepository)
	mockCustRepo := new(MockCustomerRepository)
	svc := sales.NewService(mockRepo, mockInvRepo, mockCustRepo)

	t.Run("Valid Transition", func(t *testing.T) {
		order := &domain.Order{ID: 1, Status: domain.OrderStatusDraft}
		mockRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(order, nil).Once()
		mockRepo.On("UpdateStatus", mock.Anything, int64(1), int64(1), domain.OrderStatusPendingPayment).Return(nil).Once()

		err := svc.UpdateOrderStatus(context.Background(), 1, 1, domain.OrderStatusPendingPayment)
		assert.NoError(t, err)
	})

	t.Run("Invalid Transition", func(t *testing.T) {
		order := &domain.Order{ID: 1, Status: domain.OrderStatusDelivered}
		mockRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(order, nil).Once()

		// Attempting to move from delivered back to paid is illegal
		err := svc.UpdateOrderStatus(context.Background(), 1, 1, domain.OrderStatusPaid)
		assert.ErrorIs(t, err, sales.ErrInvalidStatusTransition)
	})

	t.Run("Order Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(nil, nil).Once()

		err := svc.UpdateOrderStatus(context.Background(), 1, 1, domain.OrderStatusPendingPayment)
		assert.ErrorContains(t, err, "order not found")
	})
}

func TestSalesService_RecalculateOrderTotal(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockInvRepo := new(MockProductRepository)
	mockCustRepo := new(MockCustomerRepository)
	svc := sales.NewService(mockRepo, mockInvRepo, mockCustRepo)

	t.Run("Success", func(t *testing.T) {
		order := &domain.Order{
			ID: 1,
			Subtotal: decimal.NewFromFloat(50.0),
			DeliveryCost: decimal.NewFromFloat(0.0),
		}

		mockRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(order, nil).Once()

		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(o *domain.Order) bool {
			return o.DeliveryCost.Equal(decimal.NewFromFloat(15.0)) && o.Total.Equal(decimal.NewFromFloat(65.0))
		})).Return(nil).Once()

		err := svc.RecalculateOrderTotal(context.Background(), 1, 1, decimal.NewFromFloat(15.0))
		assert.NoError(t, err)
	})

	t.Run("Order Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(nil, nil).Once()

		err := svc.RecalculateOrderTotal(context.Background(), 1, 1, decimal.NewFromFloat(15.0))
		assert.ErrorContains(t, err, "order not found")
	})
}
