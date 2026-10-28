package chat

import (
	"context"

	"kallpa-api/internal/domain"
)

// Repository defines data operations for chat sessions and AI logs.
type Repository interface {
	GetActiveSession(ctx context.Context, storeID, customerID int64) (*domain.ChatSession, error)
	UpsertSession(ctx context.Context, session *domain.ChatSession) error
	CreateLog(ctx context.Context, log *domain.AiLog) error
}
