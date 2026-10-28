package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// TelegramIdentity maps a Telegram user ID to their profile info.
type TelegramIdentity struct {
	TelegramID   int64     `json:"telegram_id" gorm:"primaryKey"`
	FirstName    *string   `json:"first_name" gorm:"type:varchar(100)"`
	Username     *string   `json:"username" gorm:"type:varchar(100)"`
	LanguageCode *string   `json:"language_code" gorm:"type:varchar(10)"`
	CreatedAt    time.Time `json:"created_at"`
}

// Customer represents a buyer for a specific store.
type Customer struct {
	ID         int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID    int64   `json:"store_id" gorm:"index"`
	TelegramID *int64  `json:"telegram_id" gorm:"index"`
	Phone      *string `json:"phone" gorm:"type:varchar(30)"`
	Email      *string `json:"email" gorm:"type:varchar(150)"`
	FullName   *string `json:"full_name" gorm:"type:varchar(200)"`
	Address    *string `json:"address" gorm:"type:text"`

	LTV       decimal.Decimal `json:"ltv" gorm:"type:decimal(12,2);default:0"`
	Status    string          `json:"status" gorm:"type:varchar(20);default:'lead'"`
	CreatedAt time.Time       `json:"created_at"`

	TelegramIdentity *TelegramIdentity `json:"telegram_identity,omitempty" gorm:"foreignKey:TelegramID;references:TelegramID"`
}
