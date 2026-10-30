package handlers_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/telegram"
)

type MockTelegramService struct {
	mock.Mock
}

func (m *MockTelegramService) ProcessWebhook(ctx context.Context, token string, update *telegram.Update) error {
	args := m.Called(ctx, token, update)
	return args.Error(0)
}
