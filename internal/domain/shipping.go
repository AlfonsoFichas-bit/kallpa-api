package domain

import (
	"github.com/shopspring/decimal"
)

// DeliveryZone represents an available shipping area and its cost.
type DeliveryZone struct {
	ID            int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID       int64           `json:"store_id" gorm:"index"`
	Name          string          `json:"name" gorm:"type:varchar(100);not null"`
	Price         decimal.Decimal `json:"price" gorm:"type:decimal(12,2);default:0"`
	EstimatedTime *string         `json:"estimated_time" gorm:"type:varchar(50)"`
	IsActive      bool            `json:"is_active" gorm:"default:true"`
}
