package sales

import (
	"context"

	"kallpa-api/internal/domain"
)

// OrderRepository defines operations for the sales module.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, storeID, orderID int64) (*domain.Order, error)
	UpdateStatus(ctx context.Context, storeID, orderID int64, newStatus domain.OrderStatus) error
	FindLatestDraftByCustomer(ctx context.Context, storeID, customerID int64) (*domain.Order, error)
	FindByIdempotencyKey(ctx context.Context, storeID int64, key string) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
}
