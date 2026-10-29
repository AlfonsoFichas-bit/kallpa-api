package sales

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
)

// OrderPGRepository handles database operations for orders.
type OrderPGRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new instance of OrderPGRepository.
func NewOrderRepository(db *gorm.DB) *OrderPGRepository {
	return &OrderPGRepository{db: db}
}

// ListOrders retrieves orders for a store, optionally filtering by status.
func (r *OrderPGRepository) ListOrders(ctx context.Context, storeID int64, status string) ([]domain.Order, error) {
	var orders []domain.Order
	q := r.db.WithContext(ctx).Preload("Customer").Where("store_id = ?", storeID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// CreateOrder safely deduces stock and creates an order transactionally.
func (r *OrderPGRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Deduct stock for all items
		for _, item := range order.Items {
			if item.ProductVariantID == nil {
				continue
			}

			var currentStock int
			err := tx.Model(&domain.ProductVariant{}).
				Select("stock_quantity").
				Where("id = ?", *item.ProductVariantID).
				Scan(&currentStock).Error

			if err != nil {
				return err
			}

			if currentStock-item.Quantity < 0 {
				return inventory.ErrInsufficientStock
			}

			res := tx.Model(&domain.ProductVariant{}).
				Where("id = ? AND stock_quantity - ? >= 0", *item.ProductVariantID, item.Quantity).
				Update("stock_quantity", gorm.Expr("stock_quantity - ?", item.Quantity))

			if res.Error != nil {
				return res.Error
			}

			if res.RowsAffected == 0 {
				return inventory.ErrInsufficientStock
			}
		}

		// 2. Create Order
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetByID retrieves a specific order with its items and customer.
func (r *OrderPGRepository) GetByID(ctx context.Context, storeID, orderID int64) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Customer").
		Where("store_id = ? AND id = ?", storeID, orderID).
		First(&order).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// UpdateStatus changes the status of an order directly in DB.
func (r *OrderPGRepository) UpdateStatus(ctx context.Context, storeID, orderID int64, newStatus domain.OrderStatus) error {
	res := r.db.WithContext(ctx).Model(&domain.Order{}).
		Where("store_id = ? AND id = ?", storeID, orderID).
		Update("status", newStatus)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Update saves an existing order.
func (r *OrderPGRepository) Update(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// FindLatestDraftByCustomer finds the most recent draft order for a given customer.
func (r *OrderPGRepository) FindLatestDraftByCustomer(ctx context.Context, storeID, customerID int64) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND customer_id = ? AND status = ?", storeID, customerID, domain.OrderStatusDraft).
		Order("created_at DESC").
		First(&order).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// FindByIdempotencyKey checks if an order already exists with the given key.
func (r *OrderPGRepository) FindByIdempotencyKey(ctx context.Context, storeID int64, key string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Customer").
		Where("store_id = ? AND idempotency_key = ?", storeID, key).
		First(&order).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}
