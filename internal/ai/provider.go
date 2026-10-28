package ai

import (
	"context"
)

// Provider represents an interface for external AI services.
type Provider interface {
	ProcessMessage(ctx context.Context, storeID int64, message string, contextData map[string]interface{}) (*ProcessResult, error)
}

// ProcessResult is the structured response from the AI provider.
type ProcessResult struct {
	Reply        string
	Intent       string
	InputTokens  int
	OutputTokens int
	ModelUsed    string
	Score        float64
}

// DummyProvider is a mock implementation of AIProvider for testing.
type DummyProvider struct{}

// NewDummyProvider creates a new mock AI provider.
func NewDummyProvider() Provider {
	return &DummyProvider{}
}

// ProcessMessage simulates processing an AI request.
func (d *DummyProvider) ProcessMessage(ctx context.Context, storeID int64, message string, contextData map[string]interface{}) (*ProcessResult, error) {
	intent := "unknown"
	reply := "Entiendo, gracias por tu mensaje."

	// Simple keyword matching for testing intents
	if message == "Quiero 2 bolsas de quinua" {
		intent = "add_to_cart"
		reply = "Tengo quinua real disponible. ¿Deseas 2 unidades?"
	} else if message == "Si" {
		intent = "confirm_order"
		reply = "Pedido confirmado. Tu total es 100."
	}

	return &ProcessResult{
		Reply:        reply,
		Intent:       intent,
		InputTokens:  10,
		OutputTokens: 20,
		ModelUsed:    "dummy-v1",
		Score:        0.9,
	}, nil
}
