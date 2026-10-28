package integration

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
	"kallpa-api/internal/sales"
)

// TestIntegration_StockConcurrency tests Task 9.3 requirement:
// "Add concurrency tests for stock deduction and verify only one succeeds"
func TestIntegration_StockConcurrency(t *testing.T) {
	// Setup fresh database for test
	product, customer := setupStoreData(t)

	// We'll set the stock to exactly 1
	variantID := product.Variants[0].ID
	testDB.Model(&domain.ProductVariant{}).Where("id = ?", variantID).Update("stock_quantity", 1)

	// Clean orders
	testDB.Exec("DELETE FROM order_items")
	testDB.Exec("DELETE FROM orders")

	salesRepo := sales.NewOrderRepository(testDB)
	inventoryRepo := inventory.NewProductRepository(testDB)

	// Create mock customer repo specifically for the validation
	// Alternatively, rely on proper setup or bypass it by using the real PGRepository for tests if we have it
	// We'll just define a minimal manual customer fetching bypass or set it up cleanly.

	// Let's actually use the real Sales Service since we want full integration testing of the stock lock

	// First we need a stub customer repo to satisfy the interface since we aren't loading the full app graph here natively
	// Oh wait, we can just instantiate the real one!
	// In integration tests we have `testDB`.
	// We need customers repo? Let's just create a quick mock for it to avoid importing too much.

	svc := sales.NewService(salesRepo, inventoryRepo, &concurrencyMockCustomerRepo{customer: customer})
	ctx := context.Background()

	numWorkers := 15
	var successCount int32
	var failCount int32
	var wg sync.WaitGroup

	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			order := &domain.Order{
				CustomerID: customer.ID,
				Items: []domain.OrderItem{
					{ProductVariantID: &variantID, Quantity: 1},
				},
			}

			_, err := svc.CreateOrder(ctx, product.StoreID, order)
			if err != nil {
				atomic.AddInt32(&failCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful order, got %d", successCount)
	}

	if failCount != int32(numWorkers-1) {
		t.Errorf("Expected exactly %d failed orders, got %d", numWorkers-1, failCount)
	}

	// Verify final stock is 0
	var finalStock int
	testDB.Model(&domain.ProductVariant{}).Where("id = ?", variantID).Select("stock_quantity").Scan(&finalStock)

	if finalStock != 0 {
		t.Errorf("Expected final stock to be 0, got %d", finalStock)
	}
}

type concurrencyMockCustomerRepo struct {
	customer *domain.Customer
}

func (m *concurrencyMockCustomerRepo) FindByTelegramID(ctx context.Context, storeID, telegramID int64) (*domain.Customer, error) {
	return nil, nil
}
func (m *concurrencyMockCustomerRepo) UpsertByTelegramID(ctx context.Context, storeID, telegramID int64, name string) (*domain.Customer, error) {
	return nil, nil
}
func (m *concurrencyMockCustomerRepo) ListByStore(ctx context.Context, storeID int64, status string) ([]domain.Customer, error) {
	return nil, nil
}
func (m *concurrencyMockCustomerRepo) GetByID(ctx context.Context, storeID, customerID int64) (*domain.Customer, error) {
	return m.customer, nil
}
