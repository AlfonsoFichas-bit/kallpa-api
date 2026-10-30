package handlers_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"kallpa-api/internal/chat"
)

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
