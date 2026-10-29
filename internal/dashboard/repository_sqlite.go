package dashboard

import (
	"context"

	"gorm.io/gorm"

	"kallpa-api/internal/domain"
)

// Repository defines data operations for the dashboard.
type Repository interface {
	GetMetrics(ctx context.Context, storeID int64) (*MetricsResult, error)
}

// MetricsResult defines the raw database metrics result.
type MetricsResult struct {
	TotalIncome   float64
	PendingOrders int64
}

// PGRepository implements Repository via Gorm.
type PGRepository struct {
	db *gorm.DB
}

// NewRepository creates a new instance.
func NewRepository(db *gorm.DB) *PGRepository {
	return &PGRepository{db: db}
}

// GetMetrics calculates the core dashboard KPIs.
func (r *PGRepository) GetMetrics(ctx context.Context, storeID int64) (*MetricsResult, error) {
	var metrics MetricsResult

	// Get total income (sum of all paid/delivered/shipped orders)
	err := r.db.WithContext(ctx).Model(&domain.Order{}).
		Where("store_id = ? AND status IN ?", storeID, []domain.OrderStatus{domain.OrderStatusPaid, domain.OrderStatusProcessing, domain.OrderStatusShipped, domain.OrderStatusDelivered}).
		Select("COALESCE(SUM(total), 0) as total_income").
		Scan(&metrics.TotalIncome).Error

	if err != nil {
		return nil, err
	}

	// Get pending orders count
	err = r.db.WithContext(ctx).Model(&domain.Order{}).
		Where("store_id = ? AND status IN ?", storeID, []domain.OrderStatus{domain.OrderStatusPaid, domain.OrderStatusProcessing}).
		Count(&metrics.PendingOrders).Error

	if err != nil {
		return nil, err
	}

	return &metrics, nil
}
