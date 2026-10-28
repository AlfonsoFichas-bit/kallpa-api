package stores

import (
	"context"
	"fmt"

	"kallpa-api/internal/domain"
)

// Service defines business operations for stores.
type Service interface {
	GetStoreInfo(ctx context.Context, storeID int64) (*domain.Store, error)
	ListActiveStores(ctx context.Context) ([]domain.Store, error)
}

type service struct {
	repo Repository
}

// NewService creates a new store service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// GetStoreInfo returns information about a specific store.
func (s *service) GetStoreInfo(ctx context.Context, storeID int64) (*domain.Store, error) {
	store, err := s.repo.GetByID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, fmt.Errorf("store not found")
	}
	return store, nil
}

// ListActiveStores returns all active stores across the platform.
func (s *service) ListActiveStores(ctx context.Context) ([]domain.Store, error) {
	return s.repo.ListAll(ctx)
}
