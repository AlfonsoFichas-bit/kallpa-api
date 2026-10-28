package customers

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kallpa-api/internal/domain"
)

// CustomerPGRepository handles database operations for customers.
type CustomerPGRepository struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new instance of CustomerPGRepository.
func NewCustomerRepository(db *gorm.DB) *CustomerPGRepository {
	return &CustomerPGRepository{db: db}
}

// FindByTelegramID retrieves a customer for a given store by their Telegram ID.
func (r *CustomerPGRepository) FindByTelegramID(ctx context.Context, storeID, telegramID int64) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).Preload("TelegramIdentity").
		Where("store_id = ? AND telegram_id = ?", storeID, telegramID).
		First(&customer).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &customer, nil
}

// UpsertByTelegramID creates or updates a customer linked to a Telegram ID idempotently.
func (r *CustomerPGRepository) UpsertByTelegramID(ctx context.Context, storeID, telegramID int64, name string) (*domain.Customer, error) {
	var customer *domain.Customer

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Ensure TelegramIdentity exists
		var identity domain.TelegramIdentity
		err := tx.Where("telegram_id = ?", telegramID).First(&identity).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				identity = domain.TelegramIdentity{
					TelegramID: telegramID,
					FirstName:  &name,
				}
				if err := tx.Create(&identity).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// Find or Create Customer
		var cust domain.Customer
		err = tx.Where("store_id = ? AND telegram_id = ?", storeID, telegramID).First(&cust).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				cust = domain.Customer{
					StoreID:    storeID,
					TelegramID: &telegramID,
					FullName:   &name,
					Status:     "lead",
				}
				if err := tx.Create(&cust).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// Update name unconditionally for the upsert
			if cust.FullName == nil || *cust.FullName != name {
				cust.FullName = &name
				if err := tx.Save(&cust).Error; err != nil {
					return err
				}
			}
		}
		customer = &cust
		return nil
	})

	return customer, err
}

// ListByStore retrieves customers for a given store, optionally filtered by status.
func (r *CustomerPGRepository) ListByStore(ctx context.Context, storeID int64, status string) ([]domain.Customer, error) {
	var customers []domain.Customer
	q := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&customers).Error
	return customers, err
}

// GetByID retrieves a customer by ID.
func (r *CustomerPGRepository) GetByID(ctx context.Context, storeID, customerID int64) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).Where("store_id = ? AND id = ?", storeID, customerID).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &customer, nil
}
