package dashboard

import (
	"context"

	"github.com/shopspring/decimal"
)

// MetricsResponse represents the calculated dashboard metrics.
type MetricsResponse struct {
	TotalIncome   decimal.Decimal `json:"total_income"`
	PendingOrders int64           `json:"pending_orders"`
}

// Service defines business operations for the dashboard.
type Service interface {
	GetDashboardMetrics(ctx context.Context, storeID int64) (*MetricsResponse, error)
}

type service struct {
	repo Repository
}

// NewService creates a new dashboard service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// GetDashboardMetrics returns KPI metrics for a store.
func (s *service) GetDashboardMetrics(ctx context.Context, storeID int64) (*MetricsResponse, error) {
	result, err := s.repo.GetMetrics(ctx, storeID)
	if err != nil {
		return nil, err
	}

	return &MetricsResponse{
		TotalIncome:   decimal.NewFromFloat(result.TotalIncome),
		PendingOrders: result.PendingOrders,
	}, nil
}
