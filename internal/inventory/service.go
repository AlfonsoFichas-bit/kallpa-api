package inventory

import (
	"context"
	"fmt"

	"kallpa-api/internal/domain"
)

// InventoryService defines business operations for products and stock.
type InventoryService interface {
	SearchProducts(ctx context.Context, storeID int64, query string) ([]domain.Product, error)
	GetProduct(ctx context.Context, storeID, productID int64) (*domain.Product, error)
	AdjustStock(ctx context.Context, storeID, variantID int64, delta int) error
	GetLowStock(ctx context.Context, storeID int64, threshold int) ([]domain.ProductVariant, error)
}

type service struct {
	repo ProductRepository
}

// NewService creates a new InventoryService.
func NewService(repo ProductRepository) InventoryService {
	return &service{repo: repo}
}

// SearchProducts returns a list of products matching the query.
func (s *service) SearchProducts(ctx context.Context, storeID int64, query string) ([]domain.Product, error) {
	q := ProductSearchQuery{Query: query}
	return s.repo.ListProducts(ctx, storeID, q)
}

// GetProduct returns a specific product.
func (s *service) GetProduct(ctx context.Context, storeID, productID int64) (*domain.Product, error) {
	return s.repo.GetProductByID(ctx, storeID, productID)
}

// AdjustStock updates the stock of a variant ensuring it belongs to the store.
func (s *service) AdjustStock(ctx context.Context, storeID, variantID int64, delta int) error {
	variant, err := s.repo.GetVariantByID(ctx, variantID)
	if err != nil {
		return err
	}
	if variant == nil {
		return fmt.Errorf("variant not found")
	}

	// We need to fetch product to verify store ID
	product, err := s.repo.GetProductByID(ctx, storeID, variant.ProductID)
	if err != nil {
		return err
	}
	if product == nil || product.StoreID != storeID {
		return fmt.Errorf("variant does not belong to this store")
	}

	return s.repo.UpdateStock(ctx, variantID, delta)
}

// GetLowStock returns variants below the threshold.
func (s *service) GetLowStock(ctx context.Context, storeID int64, threshold int) ([]domain.ProductVariant, error) {
	return s.repo.LowStockQuery(ctx, storeID, threshold)
}
