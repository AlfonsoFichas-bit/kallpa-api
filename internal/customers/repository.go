package customers

import (
	"context"

	"kallpa-api/internal/domain"
)

// CustomerRepository defines operations for the customer module.
type CustomerRepository interface {
	FindByTelegramID(ctx context.Context, storeID int64, telegramID int64) (*domain.Customer, error)
	UpsertByTelegramID(ctx context.Context, storeID int64, telegramID int64, name string) (*domain.Customer, error)
	ListByStore(ctx context.Context, storeID int64, status string) ([]domain.Customer, error)
	GetByID(ctx context.Context, storeID int64, customerID int64) (*domain.Customer, error)
}
