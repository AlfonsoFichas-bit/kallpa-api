package handlers_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/dashboard"
)

type MockDashboardService struct {
	mock.Mock
}

func (m *MockDashboardService) GetDashboardMetrics(ctx context.Context, storeID int64) (*dashboard.MetricsResponse, error) {
	args := m.Called(ctx, storeID)
	if args.Get(0) != nil {
		return args.Get(0).(*dashboard.MetricsResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
