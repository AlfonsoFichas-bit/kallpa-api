package sales

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"kallpa-api/internal/customers"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
)

var (
	// ErrInvalidStatusTransition occurs when attempting an illegal order status change.
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// SalesService defines business operations for orders and sales.
type SalesService interface {
	ListOrders(ctx context.Context, storeID int64, status string) ([]domain.Order, error)
	CreateOrder(ctx context.Context, storeID int64, order *domain.Order) (*domain.Order, error)
	GetOrder(ctx context.Context, storeID, orderID int64) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, storeID, orderID int64, newStatus domain.OrderStatus) error
	RecalculateOrderTotal(ctx context.Context, storeID, orderID int64, deliveryCost decimal.Decimal) error
}

type service struct {
	repo          OrderRepository
	inventoryRepo inventory.ProductRepository
	customerRepo  customers.CustomerRepository
}

// NewService creates a new SalesService.
func NewService(repo OrderRepository, inventoryRepo inventory.ProductRepository, customerRepo customers.CustomerRepository) SalesService {
	return &service{
		repo:          repo,
		inventoryRepo: inventoryRepo,
		customerRepo:  customerRepo,
	}
}

// ListOrders retrieves a list of orders.
func (s *service) ListOrders(ctx context.Context, storeID int64, status string) ([]domain.Order, error) {
	return s.repo.ListOrders(ctx, storeID, status)
}

// CreateOrder validates items, calculates totals, handles idempotency, and saves.
func (s *service) CreateOrder(ctx context.Context, storeID int64, order *domain.Order) (*domain.Order, error) {
	order.StoreID = storeID

	// Validate customer belongs to this store
	customer, err := s.customerRepo.GetByID(ctx, storeID, order.CustomerID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, fmt.Errorf("customer %d not found for this store", order.CustomerID)
	}

	if order.IdempotencyKey != nil && *order.IdempotencyKey != "" {
		existing, err := s.repo.FindByIdempotencyKey(ctx, storeID, *order.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	// Calculate totals and hydrate item info
	subtotal := decimal.NewFromInt(0)
	for i := range order.Items {
		item := &order.Items[i]

		if item.ProductVariantID != nil {
			variant, err := s.inventoryRepo.GetVariantByID(ctx, *item.ProductVariantID)
			if err != nil {
				return nil, err
			}
			if variant == nil {
				return nil, fmt.Errorf("variant %d not found", *item.ProductVariantID)
			}

			if variant.Product == nil || variant.Product.StoreID != storeID {
				return nil, fmt.Errorf("variant %d does not belong to this store", *item.ProductVariantID)
			}

			// Compute effective unit price
			unitPrice := variant.Product.BasePrice.Add(variant.AdditionalPrice)
			item.UnitPrice = unitPrice

			if variant.Product != nil {
				item.ProductName = &variant.Product.Name
			}

			var variantName string
			if variant.Size != nil {
				variantName += *variant.Size
			}
			if variant.Color != nil {
				if variantName != "" {
					variantName += " - "
				}
				variantName += *variant.Color
			}
			item.VariantName = &variantName
		}

		item.TotalLine = item.UnitPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
		subtotal = subtotal.Add(item.TotalLine)
	}

	order.Subtotal = subtotal
	order.Total = subtotal.Add(order.DeliveryCost)

	// Delegate transactional creation + stock deduction to repository
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// GetOrder returns a single order.
func (s *service) GetOrder(ctx context.Context, storeID, orderID int64) (*domain.Order, error) {
	return s.repo.GetByID(ctx, storeID, orderID)
}

// RecalculateOrderTotal updates an order's delivery cost and total.
func (s *service) RecalculateOrderTotal(ctx context.Context, storeID, orderID int64, deliveryCost decimal.Decimal) error {
	order, err := s.repo.GetByID(ctx, storeID, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	order.DeliveryCost = deliveryCost
	order.Total = order.Subtotal.Add(deliveryCost)

	return s.repo.Update(ctx, order)
}

// UpdateOrderStatus transitions an order to a new status.
func (s *service) UpdateOrderStatus(ctx context.Context, storeID, orderID int64, newStatus domain.OrderStatus) error {
	order, err := s.repo.GetByID(ctx, storeID, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	if !isValidTransition(order.Status, newStatus) {
		return ErrInvalidStatusTransition
	}

	return s.repo.UpdateStatus(ctx, storeID, orderID, newStatus)
}

func isValidTransition(current, next domain.OrderStatus) bool {
	// Status transitions: draft -> pending_payment -> paid -> processing -> shipped -> delivered
	// cancelled only from non-terminal (draft, pending_payment, paid, processing)
	// Terminal states: delivered, cancelled, returned

	if current == next {
		return true // Idempotent updates are fine
	}

	switch current {
	case domain.OrderStatusDraft:
		return next == domain.OrderStatusPendingPayment || next == domain.OrderStatusCancelled
	case domain.OrderStatusPendingPayment:
		return next == domain.OrderStatusPaid || next == domain.OrderStatusCancelled
	case domain.OrderStatusPaid:
		return next == domain.OrderStatusProcessing || next == domain.OrderStatusCancelled
	case domain.OrderStatusProcessing:
		return next == domain.OrderStatusShipped || next == domain.OrderStatusCancelled
	case domain.OrderStatusShipped:
		return next == domain.OrderStatusDelivered || next == domain.OrderStatusReturned
	case domain.OrderStatusDelivered, domain.OrderStatusCancelled, domain.OrderStatusReturned:
		return false // Terminal states
	}

	return false
}
