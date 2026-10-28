package inventory

import (
	"context"

	"kallpa-api/internal/domain"
)

// ProductSearchQuery encapsulates parameters for searching products.
type ProductSearchQuery struct {
	Query string
}

// ProductRepository defines operations for the inventory module.
type ProductRepository interface {
	ListProducts(ctx context.Context, storeID int64, query ProductSearchQuery) ([]domain.Product, error)
	GetProductByID(ctx context.Context, storeID int64, productID int64) (*domain.Product, error)
	GetVariantByID(ctx context.Context, variantID int64) (*domain.ProductVariant, error)
	UpdateStock(ctx context.Context, variantID int64, delta int) error
	LowStockQuery(ctx context.Context, storeID int64, threshold int) ([]domain.ProductVariant, error)
}
