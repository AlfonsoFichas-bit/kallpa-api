package services_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/ai"
	"kallpa-api/internal/domain"
)

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) GetActiveSession(ctx context.Context, storeID, customerID int64) (*domain.ChatSession, error) {
	args := m.Called(ctx, storeID, customerID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.ChatSession), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChatRepository) UpsertSession(ctx context.Context, session *domain.ChatSession) error {
	args := m.Called(ctx, session)
	if session != nil && session.ID == 0 && args.Error(0) == nil {
		session.ID = 1
	}
	return args.Error(0)
}

func (m *MockChatRepository) CreateLog(ctx context.Context, log *domain.AiLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

type MockAIProvider struct {
	mock.Mock
}

func (m *MockAIProvider) ProcessMessage(ctx context.Context, storeID int64, message string, contextData map[string]interface{}) (*ai.ProcessResult, error) {
	args := m.Called(ctx, storeID, message, contextData)
	if args.Get(0) != nil {
		return args.Get(0).(*ai.ProcessResult), args.Error(1)
	}
	return nil, args.Error(1)
}
