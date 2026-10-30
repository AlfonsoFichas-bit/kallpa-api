package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/domain"
	"kallpa-api/internal/shipping"
)

func TestShippingService_ListStoreZones(t *testing.T) {
	mockRepo := new(MockDeliveryZoneRepository)
	mockSales := new(MockSalesService)
	svc := shipping.NewService(mockRepo, mockSales)

	t.Run("Success with Virtual Pickup", func(t *testing.T) {
		zones := []domain.DeliveryZone{
			{ID: 1, Name: "Downtown", Price: decimal.NewFromFloat(15.0)},
		}

		mockRepo.On("ListZones", mock.Anything, int64(1)).Return(zones, nil).Once()

		result, err := svc.ListStoreZones(context.Background(), 1)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(-1), result[0].ID) // Pickup zone added automatically
		assert.Equal(t, "Recojo en tienda", result[0].Name)
		assert.Equal(t, int64(1), result[1].ID) // Actual DB zone
	})

	t.Run("DB Error", func(t *testing.T) {
		mockRepo.On("ListZones", mock.Anything, int64(1)).Return(nil, errors.New("db error")).Once()

		result, err := svc.ListStoreZones(context.Background(), 1)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestShippingService_AssignZoneToOrder(t *testing.T) {
	mockRepo := new(MockDeliveryZoneRepository)
	mockSales := new(MockSalesService)
	svc := shipping.NewService(mockRepo, mockSales)

	t.Run("Assign Virtual Pickup", func(t *testing.T) {
		mockSales.On("RecalculateOrderTotal", mock.Anything, int64(1), int64(2), mock.MatchedBy(func(d decimal.Decimal) bool {
			return d.Equal(decimal.NewFromInt(0))
		})).Return(nil).Once()

		err := svc.AssignZoneToOrder(context.Background(), 1, 2, -1) // -1 is Pickup
		assert.NoError(t, err)
	})

	t.Run("Assign Valid Zone", func(t *testing.T) {
		zone := &domain.DeliveryZone{ID: 10, IsActive: true, Price: decimal.NewFromFloat(20.5)}
		mockRepo.On("GetByID", mock.Anything, int64(1), int64(10)).Return(zone, nil).Once()

		mockSales.On("RecalculateOrderTotal", mock.Anything, int64(1), int64(2), mock.MatchedBy(func(d decimal.Decimal) bool {
			return d.Equal(decimal.NewFromFloat(20.5))
		})).Return(nil).Once()

		err := svc.AssignZoneToOrder(context.Background(), 1, 2, 10)
		assert.NoError(t, err)
	})

	t.Run("Zone Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, int64(1), int64(10)).Return(nil, nil).Once()

		err := svc.AssignZoneToOrder(context.Background(), 1, 2, 10)
		assert.ErrorIs(t, err, shipping.ErrZoneNotFound)
	})

	t.Run("Zone Inactive", func(t *testing.T) {
		zone := &domain.DeliveryZone{ID: 10, IsActive: false, Price: decimal.NewFromFloat(20.5)}
		mockRepo.On("GetByID", mock.Anything, int64(1), int64(10)).Return(zone, nil).Once()

		err := svc.AssignZoneToOrder(context.Background(), 1, 2, 10)
		assert.ErrorIs(t, err, shipping.ErrZoneNotFound)
	})
}

func TestShippingService_CreateZone(t *testing.T) {
	mockRepo := new(MockDeliveryZoneRepository)
	mockSales := new(MockSalesService)
	svc := shipping.NewService(mockRepo, mockSales)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("CreateZone", mock.Anything, mock.AnythingOfType("*domain.DeliveryZone")).Return(nil).Once()

		zone, err := svc.CreateZone(context.Background(), 1, "North", decimal.NewFromFloat(15.0))
		assert.NoError(t, err)
		assert.NotNil(t, zone)
		assert.Equal(t, "North", zone.Name)
		assert.True(t, zone.IsActive)
	})
}

func TestShippingService_DeleteZone(t *testing.T) {
	mockRepo := new(MockDeliveryZoneRepository)
	mockSales := new(MockSalesService)
	svc := shipping.NewService(mockRepo, mockSales)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("DeleteZone", mock.Anything, int64(1), int64(2)).Return(nil).Once()

		err := svc.DeleteZone(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("DeleteZone", mock.Anything, int64(1), int64(2)).Return(errors.New("record not found")).Once()

		err := svc.DeleteZone(context.Background(), 1, 2)
		assert.ErrorIs(t, err, shipping.ErrZoneNotFound)
	})
}
