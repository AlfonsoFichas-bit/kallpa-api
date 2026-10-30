package services_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/chat"
	"kallpa-api/internal/domain"
)

type MockTelegramClient struct {
	mock.Mock
}

func (m *MockTelegramClient) SendMessage(ctx context.Context, botToken string, chatID int64, text string) error {
	args := m.Called(ctx, botToken, chatID, text)
	return args.Error(0)
}

func (m *MockTelegramClient) GetFileURL(ctx context.Context, botToken, fileID string) (string, error) {
	args := m.Called(ctx, botToken, fileID)
	return args.String(0), args.Error(1)
}

type MockStoreService struct {
	mock.Mock
}

func (m *MockStoreService) GetStoreInfo(ctx context.Context, storeID int64) (*domain.Store, error) {
	args := m.Called(ctx, storeID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Store), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStoreService) ListActiveStores(ctx context.Context) ([]domain.Store, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Store), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) ProcessMessage(ctx context.Context, req *chat.MessageRequest) (*chat.MessageResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) != nil {
		return args.Get(0).(*chat.MessageResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
