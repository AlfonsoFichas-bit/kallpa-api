package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderStatus defines the lifecycle states of an order.
type OrderStatus string

const (
	// OrderStatusDraft represents an order that is still being built.
	OrderStatusDraft OrderStatus = "draft"
	// OrderStatusPendingPayment represents an order awaiting customer payment.
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	// OrderStatusPaid represents a fully paid order.
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusProcessing represents an order being prepared for shipping.
	OrderStatusProcessing OrderStatus = "processing"
	// OrderStatusShipped represents an order that has left the store.
	OrderStatusShipped OrderStatus = "shipped"
	// OrderStatusDelivered represents a successfully completed order.
	OrderStatusDelivered OrderStatus = "delivered"
	// OrderStatusCancelled represents an order aborted by the store or customer.
	OrderStatusCancelled OrderStatus = "cancelled"
	// OrderStatusReturned represents an order returned by the customer.
	OrderStatusReturned OrderStatus = "returned"
)

// PaymentMethod specifies how the order was paid.
type PaymentMethod string

const (
	// PaymentMethodQRTransfer implies payment via bank QR.
	PaymentMethodQRTransfer PaymentMethod = "qr_transfer"
	// PaymentMethodCashOnDelivery implies payment will be collected at delivery.
	PaymentMethodCashOnDelivery PaymentMethod = "cash_on_delivery"
	// PaymentMethodCardLink implies payment via an external credit card gateway link.
	PaymentMethodCardLink PaymentMethod = "card_link"
)

// Order represents a customer's purchase.
type Order struct {
	ID            int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID       int64          `json:"store_id" gorm:"index"`
	CustomerID    int64          `json:"customer_id" gorm:"index"`
	Status        OrderStatus    `json:"status" gorm:"type:varchar(50);default:'draft'"`
	PaymentMethod *PaymentMethod `json:"payment_method" gorm:"type:varchar(50)"`

	Subtotal     decimal.Decimal `json:"subtotal" gorm:"type:decimal(12,2);not null;default:0"`
	DeliveryCost decimal.Decimal `json:"delivery_cost" gorm:"type:decimal(12,2);default:0"`
	Total        decimal.Decimal `json:"total" gorm:"type:decimal(12,2);not null;default:0"`

	DeliveryZoneID  *int64  `json:"delivery_zone_id" gorm:"index"`
	ShippingAddress *string `json:"shipping_address" gorm:"type:text"`

	PaymentProofURL *string `json:"payment_proof_url" gorm:"type:text"`
	Notes           *string `json:"notes" gorm:"type:text"`

	IdempotencyKey *string `json:"idempotency_key" gorm:"type:varchar(255);uniqueIndex"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Items    []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	Customer *Customer   `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
}

// OrderItem represents a line item within an order.
type OrderItem struct {
	ID               int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderID          int64           `json:"order_id" gorm:"index"`
	ProductVariantID *int64          `json:"product_variant_id" gorm:"index"`
	ProductName      *string         `json:"product_name" gorm:"type:varchar(200)"`
	VariantName      *string         `json:"variant_name" gorm:"type:varchar(100)"`
	Quantity         int             `json:"quantity" gorm:"not null"`
	UnitPrice        decimal.Decimal `json:"unit_price" gorm:"type:decimal(12,2);not null"`
	TotalLine        decimal.Decimal `json:"total_line" gorm:"type:decimal(12,2);not null"`
}
