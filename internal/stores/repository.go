package stores

import (
	"context"

	"kallpa-api/internal/domain"
)

// Repository defines data operations for stores.
type Repository interface {
	GetByID(ctx context.Context, storeID int64) (*domain.Store, error)
	ListAll(ctx context.Context) ([]domain.Store, error)
}
