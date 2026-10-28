package customers

import (
	"context"

	"kallpa-api/internal/domain"
)

// CustomerService defines business operations for customers.
type CustomerService interface {
	UpsertCustomerFromTelegram(ctx context.Context, storeID int64, telegramID int64, name string) (*domain.Customer, error)
	GetCustomers(ctx context.Context, storeID int64, status string) ([]domain.Customer, error)
	GetCustomer(ctx context.Context, storeID int64, id int64) (*domain.Customer, error)
}

type service struct {
	repo CustomerRepository
}

// NewService creates a new CustomerService.
func NewService(repo CustomerRepository) CustomerService {
	return &service{repo: repo}
}

// UpsertCustomerFromTelegram creates or retrieves a customer idempotently.
func (s *service) UpsertCustomerFromTelegram(ctx context.Context, storeID int64, telegramID int64, name string) (*domain.Customer, error) {
	return s.repo.UpsertByTelegramID(ctx, storeID, telegramID, name)
}

// GetCustomers returns a list of customers, optionally filtered.
func (s *service) GetCustomers(ctx context.Context, storeID int64, status string) ([]domain.Customer, error) {
	return s.repo.ListByStore(ctx, storeID, status)
}

// GetCustomer returns a single customer by ID.
func (s *service) GetCustomer(ctx context.Context, storeID int64, id int64) (*domain.Customer, error) {
	return s.repo.GetByID(ctx, storeID, id)
}
