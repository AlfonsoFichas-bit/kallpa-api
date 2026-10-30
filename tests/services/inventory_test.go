package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
)

func TestInventoryService_SearchProducts(t *testing.T) {
	mockRepo := new(MockProductRepository)
	svc := inventory.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		expectedProducts := []domain.Product{
			{ID: 1, StoreID: 1, Name: "Shoes"},
		}

		mockRepo.On("ListProducts", mock.Anything, int64(1), inventory.ProductSearchQuery{Query: "shoe"}).Return(expectedProducts, nil).Once()

		products, err := svc.SearchProducts(context.Background(), 1, "shoe")
		assert.NoError(t, err)
		assert.Len(t, products, 1)
		assert.Equal(t, "Shoes", products[0].Name)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo.On("ListProducts", mock.Anything, int64(1), inventory.ProductSearchQuery{Query: "error"}).Return(nil, errors.New("db error")).Once()

		products, err := svc.SearchProducts(context.Background(), 1, "error")
		assert.Error(t, err)
		assert.Nil(t, products)
	})
}

func TestInventoryService_GetProduct(t *testing.T) {
	mockRepo := new(MockProductRepository)
	svc := inventory.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		expectedProduct := &domain.Product{ID: 1, StoreID: 1, Name: "Shoes"}

		mockRepo.On("GetProductByID", mock.Anything, int64(1), int64(1)).Return(expectedProduct, nil).Once()

		product, err := svc.GetProduct(context.Background(), 1, 1)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, "Shoes", product.Name)
	})
}

func TestInventoryService_AdjustStock(t *testing.T) {
	mockRepo := new(MockProductRepository)
	svc := inventory.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		variant := &domain.ProductVariant{ID: 1, ProductID: 2}
		product := &domain.Product{ID: 2, StoreID: 1}

		mockRepo.On("GetVariantByID", mock.Anything, int64(1)).Return(variant, nil).Once()
		mockRepo.On("GetProductByID", mock.Anything, int64(1), int64(2)).Return(product, nil).Once()
		mockRepo.On("UpdateStock", mock.Anything, int64(1), -5).Return(nil).Once()

		err := svc.AdjustStock(context.Background(), 1, 1, -5)
		assert.NoError(t, err)
	})

	t.Run("Variant Not Found", func(t *testing.T) {
		mockRepo.On("GetVariantByID", mock.Anything, int64(1)).Return(nil, nil).Once()

		err := svc.AdjustStock(context.Background(), 1, 1, -5)
		assert.ErrorContains(t, err, "variant not found")
	})

	t.Run("Product Belongs To Another Store", func(t *testing.T) {
		variant := &domain.ProductVariant{ID: 1, ProductID: 2}
		product := &domain.Product{ID: 2, StoreID: 99} // Different store

		mockRepo.On("GetVariantByID", mock.Anything, int64(1)).Return(variant, nil).Once()
		mockRepo.On("GetProductByID", mock.Anything, int64(1), int64(2)).Return(product, nil).Once()

		err := svc.AdjustStock(context.Background(), 1, 1, -5)
		assert.ErrorContains(t, err, "variant does not belong to this store")
	})

	t.Run("Product Not Found", func(t *testing.T) {
		variant := &domain.ProductVariant{ID: 1, ProductID: 2}

		mockRepo.On("GetVariantByID", mock.Anything, int64(1)).Return(variant, nil).Once()
		mockRepo.On("GetProductByID", mock.Anything, int64(1), int64(2)).Return(nil, nil).Once()

		err := svc.AdjustStock(context.Background(), 1, 1, -5)
		assert.ErrorContains(t, err, "variant does not belong to this store")
	})
}

func TestInventoryService_GetLowStock(t *testing.T) {
	mockRepo := new(MockProductRepository)
	svc := inventory.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		expectedVariants := []domain.ProductVariant{
			{ID: 1, StockQuantity: 2},
		}

		mockRepo.On("LowStockQuery", mock.Anything, int64(1), 5).Return(expectedVariants, nil).Once()

		variants, err := svc.GetLowStock(context.Background(), 1, 5)
		assert.NoError(t, err)
		assert.Len(t, variants, 1)
	})
}

func TestInventoryService_ToggleProductStatus(t *testing.T) {
	mockRepo := new(MockProductRepository)
	svc := inventory.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("ToggleProductStatus", mock.Anything, int64(1), int64(1)).Return(nil).Once()

		err := svc.ToggleProductStatus(context.Background(), 1, 1)
		assert.NoError(t, err)
	})
}
