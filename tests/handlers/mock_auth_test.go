package handlers_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/auth"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*auth.TokenResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.TokenResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenStr string) (*auth.CustomClaims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.CustomClaims), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) Refresh(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.TokenResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
