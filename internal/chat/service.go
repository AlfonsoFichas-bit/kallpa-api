package chat

import (
	"context"
	"encoding/json"

	"kallpa-api/internal/ai"
	"kallpa-api/internal/customers"
	"kallpa-api/internal/domain"
)

// MessageRequest is the incoming conversational payload.
type MessageRequest struct {
	StoreID        int64  `json:"store_id" validate:"required"`
	ChatID         string `json:"chat_id" validate:"required"`
	TelegramUserID int64  `json:"telegram_user_id" validate:"required"`
	UserName       string `json:"user_name" validate:"required"`
	Message        string `json:"message" validate:"required"`
}

// MessageResponse is the outbound conversational payload.
type MessageResponse struct {
	Reply                string  `json:"reply"`
	Intent               string  `json:"intent"`
	OrderID              *int64  `json:"order_id,omitempty"`
	RequiresConfirmation bool    `json:"requires_confirmation"`
	QRURL                *string `json:"qr_url,omitempty"`
}

// Service defines business operations for conversational flow.
type Service interface {
	ProcessMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error)
}

type service struct {
	repo        Repository
	customerSvc customers.CustomerService
	aiProvider  ai.Provider
}

// NewService creates a new ChatService.
func NewService(repo Repository, customerSvc customers.CustomerService, aiProvider ai.Provider) Service {
	return &service{
		repo:        repo,
		customerSvc: customerSvc,
		aiProvider:  aiProvider,
	}
}

// ProcessMessage orchestrates the conversation flow.
func (s *service) ProcessMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	// 1. Idempotently get/create customer from Telegram info
	customer, err := s.customerSvc.UpsertCustomerFromTelegram(ctx, req.StoreID, req.TelegramUserID, req.UserName)
	if err != nil {
		return nil, err
	}

	// 2. Fetch or create active ChatSession
	session, err := s.repo.GetActiveSession(ctx, req.StoreID, customer.ID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		session = &domain.ChatSession{
			StoreID:    req.StoreID,
			CustomerID: customer.ID,
			IsActive:   true,
		}
	}

	// Deserialize context JSON for AI prompt
	var contextData map[string]interface{}
	if len(session.ContextData) > 0 {
		_ = json.Unmarshal(session.ContextData, &contextData)
	}

	// 3. Call AI Provider
	aiResult, err := s.aiProvider.ProcessMessage(ctx, req.StoreID, req.Message, contextData)
	if err != nil {
		return nil, err
	}

	// 4. Log interaction
	aiLog := &domain.AiLog{
		StoreID:        &req.StoreID,
		InputTokens:    &aiResult.InputTokens,
		OutputTokens:   &aiResult.OutputTokens,
		ModelUsed:      &aiResult.ModelUsed,
		UserMessage:    &req.Message,
		AiResponse:     &aiResult.Reply,
		SentimentScore: &aiResult.Score,
	}

	// We'll link the session ID after saving the session if it's new

	// 5. Update session context/flow if needed (e.g., track intent state)
	intent := aiResult.Intent
	session.CurrentStep = &intent
	if err := s.repo.UpsertSession(ctx, session); err != nil {
		return nil, err
	}

	aiLog.ChatSessionID = &session.ID
	if err := s.repo.CreateLog(ctx, aiLog); err != nil {
		// Log the error but don't fail the user interaction if logging fails
		// (Assuming a logger is available, or we just silently fail for now per previous setup)
	}

	// 6. Map to Domain Response Contract
	// Note: In a complete implementation, specific intents (like "add_to_cart")
	// would trigger calls to the sales/inventory service here before returning.
	// We keep it strictly to the contract for Phase 4.

	res := &MessageResponse{
		Reply:  aiResult.Reply,
		Intent: aiResult.Intent,
	}

	// Mocking behavior for specific intents just to fulfill contract fields conditionally
	if intent == "add_to_cart" {
		res.RequiresConfirmation = true
	} else if intent == "confirm_order" {
		oid := int64(42)
		res.OrderID = &oid
		url := "https://example.com/qr"
		res.QRURL = &url
	}

	return res, nil
}
