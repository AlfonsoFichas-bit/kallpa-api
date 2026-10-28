package domain

import (
	"time"

	"gorm.io/datatypes"
)

// Store represents a merchant utilizing the Kallpa platform.
type Store struct {
	ID              int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string         `json:"name" gorm:"type:varchar(200);not null"`
	Slug            string         `json:"slug" gorm:"type:varchar(100);unique;not null"`
	BotToken        *string        `json:"bot_token" gorm:"type:varchar(255)"`
	Currency        string         `json:"currency" gorm:"type:varchar(3);default:'BOB'"`
	QRImageURL      *string        `json:"qr_image_url" gorm:"type:text"`
	OwnerTelegramID *int64         `json:"owner_telegram_id" gorm:"index"`
	Settings        datatypes.JSON `json:"settings" gorm:"type:json"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
