package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Product represents a sellable item in a store.
type Product struct {
	ID          int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID     int64           `json:"store_id" gorm:"index"`
	Name        string          `json:"name" gorm:"type:varchar(200);not null"`
	Description *string         `json:"description" gorm:"type:text"`
	BasePrice   decimal.Decimal `json:"base_price" gorm:"type:decimal(12,2);not null"`
	Category    *string         `json:"category" gorm:"type:varchar(100)"`
	Status      string          `json:"status" gorm:"type:varchar(20);default:'active'"`
	ImageURL    *string         `json:"image_url" gorm:"type:text"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	Variants []ProductVariant `json:"variants,omitempty" gorm:"foreignKey:ProductID"`
}

// ProductVariant represents a specific version (SKU/size/color) of a Product.
type ProductVariant struct {
	ID              int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID       int64           `json:"product_id" gorm:"index"`
	SKU             *string         `json:"sku" gorm:"type:varchar(50)"`
	Size            *string         `json:"size" gorm:"type:varchar(50)"`
	Color           *string         `json:"color" gorm:"type:varchar(50)"`
	AdditionalPrice decimal.Decimal `json:"additional_price" gorm:"type:decimal(12,2);default:0"`
	StockQuantity   int             `json:"stock_quantity" gorm:"default:0"`
	CreatedAt       time.Time       `json:"created_at"`

	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}
