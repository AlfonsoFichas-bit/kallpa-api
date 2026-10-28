package domain

import (
	"time"
)

// NotificationType defines the reason for the notification.
type NotificationType string

const (
	// NotificationTypeSale implies a new sale was completed.
	NotificationTypeSale NotificationType = "sale"
	// NotificationTypeAlert implies an administrative alert.
	NotificationTypeAlert NotificationType = "alert"
	// NotificationTypeStockLow implies inventory requires attention.
	NotificationTypeStockLow NotificationType = "stock_low"
)

// Notification represents a system or dashboard alert for a store.
type Notification struct {
	ID        int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID   int64            `json:"store_id" gorm:"index"`
	OrderID   *int64           `json:"order_id" gorm:"index"`
	Type      NotificationType `json:"type" gorm:"type:varchar(50)"`
	Message   string           `json:"message" gorm:"type:text"`
	IsRead    bool             `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time        `json:"created_at"`
}
