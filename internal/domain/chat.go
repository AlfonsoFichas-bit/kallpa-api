package domain

import (
	"time"

	"gorm.io/datatypes"
)

// ChatSession represents an active conversation with a customer.
type ChatSession struct {
	ID         int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID    int64 `json:"store_id" gorm:"index"`
	CustomerID int64 `json:"customer_id" gorm:"index"`

	CurrentFlow *string        `json:"current_flow" gorm:"type:varchar(50)"`
	CurrentStep *string        `json:"current_step" gorm:"type:varchar(50)"`
	ContextData datatypes.JSON `json:"context_data" gorm:"type:json"`

	LastInteractionAt time.Time `json:"last_interaction_at" gorm:"autoUpdateTime"`
	IsActive          bool      `json:"is_active" gorm:"default:true"`
}

// AiLog records a single interaction with the AI provider.
type AiLog struct {
	ID             int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID        *int64    `json:"store_id" gorm:"index"`
	ChatSessionID  *int64    `json:"chat_session_id" gorm:"index"`
	InputTokens    *int      `json:"input_tokens"`
	OutputTokens   *int      `json:"output_tokens"`
	ModelUsed      *string   `json:"model_used" gorm:"type:varchar(50)"`
	UserMessage    *string   `json:"user_message" gorm:"type:text"`
	AiResponse     *string   `json:"ai_response" gorm:"type:text"`
	SentimentScore *float64  `json:"sentiment_score" gorm:"type:decimal(3,2)"`
	CreatedAt      time.Time `json:"created_at"`
}
