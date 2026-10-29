package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"kallpa-api/internal/domain"
)

// ErrInsufficientStock is returned when a stock update would result in negative quantity.
var ErrInsufficientStock = errors.New("insufficient stock")

// ProductPGRepository handles database operations for inventory.
type ProductPGRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new instance of ProductPGRepository.
func NewProductRepository(db *gorm.DB) *ProductPGRepository {
	return &ProductPGRepository{db: db}
}

// ListProducts retrieves active products for a store, optionally filtering by search query.
func (r *ProductPGRepository) ListProducts(ctx context.Context, storeID int64, search ProductSearchQuery) ([]domain.Product, error) {
	var products []domain.Product
	q := r.db.WithContext(ctx).Preload("Variants").Where("store_id = ? AND status = ?", storeID, "active")

	if search.Query != "" {
		likeQuery := "%" + strings.ToLower(search.Query) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(category) LIKE ?", likeQuery, likeQuery, likeQuery)
	}

	err := q.Find(&products).Error
	return products, err
}

// GetProductByID retrieves a single product with its variants for a given store.
func (r *ProductPGRepository) GetProductByID(ctx context.Context, storeID, productID int64) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).Preload("Variants").Where("store_id = ? AND id = ?", storeID, productID).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// GetVariantByID retrieves a specific product variant.
func (r *ProductPGRepository) GetVariantByID(ctx context.Context, variantID int64) (*domain.ProductVariant, error) {
	var variant domain.ProductVariant
	err := r.db.WithContext(ctx).Preload("Product").First(&variant, variantID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &variant, nil
}

// UpdateStock changes the stock of a variant safely, returning an error if it goes negative.
func (r *ProductPGRepository) UpdateStock(ctx context.Context, variantID int64, delta int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var currentStock int

		// Lock the row (In SQLite, FOR UPDATE is generally unsupported or locks the whole DB,
		// but GORM abstract handles it where possible. For SQLite, transaction serialization protects this).
		err := tx.Model(&domain.ProductVariant{}).
			Select("stock_quantity").
			Where("id = ?", variantID).
			Scan(&currentStock).Error

		if err != nil {
			return err
		}

		if currentStock+delta < 0 {
			return ErrInsufficientStock
		}

		res := tx.Model(&domain.ProductVariant{}).
			Where("id = ? AND stock_quantity + ? >= 0", variantID, delta).
			Update("stock_quantity", gorm.Expr("stock_quantity + ?", delta))

		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			// Check if the variant actually exists, if rows affected is 0 it could be a race condition.
			var count int64
			tx.Model(&domain.ProductVariant{}).Where("id = ?", variantID).Count(&count)
			if count == 0 {
				return fmt.Errorf("variant not found")
			}
			return ErrInsufficientStock
		}

		return nil
	})
}

// LowStockQuery returns variants for a store that are at or below the given threshold.
func (r *ProductPGRepository) LowStockQuery(ctx context.Context, storeID int64, threshold int) ([]domain.ProductVariant, error) {
	var variants []domain.ProductVariant
	err := r.db.WithContext(ctx).
		Joins("JOIN products ON products.id = product_variants.product_id").
		Where("products.store_id = ? AND product_variants.stock_quantity <= ?", storeID, threshold).
		Find(&variants).Error
	return variants, err
}

// ToggleProductStatus flips a product's active status.
func (r *ProductPGRepository) ToggleProductStatus(ctx context.Context, storeID, productID int64) error {
	var product domain.Product
	err := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", productID, storeID).First(&product).Error
	if err != nil {
		return err
	}

	newStatus := "active"
	if product.Status == "active" {
		newStatus = "inactive"
	}

	return r.db.WithContext(ctx).Model(&product).Update("status", newStatus).Error
}
