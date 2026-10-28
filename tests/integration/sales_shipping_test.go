package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"

	"kallpa-api/internal/domain"
	"kallpa-api/internal/sales"
	"kallpa-api/internal/shipping"
)

func setupSalesShippingData(t *testing.T) (*domain.Product, *domain.Customer, *domain.DeliveryZone) {
	testDB.Exec("DELETE FROM order_items")
	testDB.Exec("DELETE FROM orders")
	testDB.Exec("DELETE FROM delivery_zones")

	product, customer := setupStoreData(t) // Reuse helper from previous tests

	zone := &domain.DeliveryZone{
		StoreID:  1,
		Name:     "Centro",
		Price:    decimal.NewFromInt(10),
		IsActive: true,
	}
	if err := testDB.Create(zone).Error; err != nil {
		t.Fatalf("failed to create delivery zone: %v", err)
	}

	return product, customer, zone
}

func TestIntegration_SalesShippingFlow(t *testing.T) {
	product, customer, zone := setupSalesShippingData(t)
	token := getAdminAuthToken(t)

	// 1. List Delivery Zones
	req := httptest.NewRequest(http.MethodGet, "/v1/delivery-zones", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for delivery zones list, got %d", rec.Code)
	}
	var zones []domain.DeliveryZone
	json.Unmarshal(rec.Body.Bytes(), &zones)
	if len(zones) != 2 {
		t.Errorf("expected 2 zones (1 virtual, 1 real), got %d", len(zones))
	}

	// 2. Create Order
	variantID := product.Variants[0].ID
	createReq := sales.CreateOrderRequest{
		CustomerID: customer.ID,
		Items: []sales.CreateOrderItem{
			{ProductVariantID: variantID, Quantity: 2}, // Assuming the handler currently relies on unit prices being set; wait, our handler doesn't set UnitPrice!
		},
	}

	// Because our handler currently doesn't fetch UnitPrice from the DB and maps it directly,
	// for the integration test to pass we might need to modify the handler to fetch it,
	// or we mock the request to not care about totals right now.
	// Actually, the requirements explicitly say: "Compute subtotal = sum(line totals), total = subtotal + delivery_cost."
	// Let's rely on what we implemented and test the stock deduction.
	body, _ := json.Marshal(createReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", "test-idem-key")
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for order, got %d. Body: %s", rec.Code, rec.Body.String())
	}
	var order domain.Order
	json.Unmarshal(rec.Body.Bytes(), &order)
	if order.ID == 0 {
		t.Fatalf("expected order to be created")
	}

	// Verify Stock Deduction (Initial stock was 10, bought 2 -> expected 8)
	var updatedVariant domain.ProductVariant
	testDB.First(&updatedVariant, variantID)
	if updatedVariant.StockQuantity != 8 {
		t.Errorf("expected stock to be deducted to 8, got %d", updatedVariant.StockQuantity)
	}

	// 3. Test Idempotency (Duplicate Key)
	req = httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", "test-idem-key")
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for duplicate idempotency key, got %d", rec.Code)
	}
	var duplicateOrder domain.Order
	json.Unmarshal(rec.Body.Bytes(), &duplicateOrder)
	if duplicateOrder.ID != order.ID {
		t.Errorf("expected duplicate order to return same ID %d, got %d", order.ID, duplicateOrder.ID)
	}

	// Verify stock was not deducted again
	testDB.First(&updatedVariant, variantID)
	if updatedVariant.StockQuantity != 8 {
		t.Errorf("expected stock to remain 8 for idempotent request, got %d", updatedVariant.StockQuantity)
	}

	// 4. Assign Shipping Zone
	assignReq := shipping.AssignZoneRequest{
		ZoneID: zone.ID,
	}
	assignBody, _ := json.Marshal(assignReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/orders/"+strconv.FormatInt(order.ID, 10)+"/shipping", bytes.NewReader(assignBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for shipping assignment, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Get Order and Verify Totals Recalculation
	req = httptest.NewRequest(http.MethodGet, "/v1/orders/"+strconv.FormatInt(order.ID, 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for get order, got %d", rec.Code)
	}
	var fetchedOrder domain.Order
	json.Unmarshal(rec.Body.Bytes(), &fetchedOrder)

	if !fetchedOrder.DeliveryCost.Equal(zone.Price) {
		t.Errorf("expected delivery cost %s, got %s", zone.Price, fetchedOrder.DeliveryCost)
	}

	// We now hydrate unit prices properly in the sales service.
	// We bought 2 items of Test Quinoa. Base price is 50. TotalLine = 2 * 50 = 100.
	// Subtotal is 100. Total = 100 + 10 (Delivery) = 110.
	expectedTotal := decimal.NewFromInt(110)
	if !fetchedOrder.Total.Equal(expectedTotal) {
		t.Errorf("expected total %s, got %s", expectedTotal, fetchedOrder.Total)
	}

	// 5. Update Status
	statusReq := sales.UpdateOrderStatusRequest{Status: string(domain.OrderStatusPendingPayment)}
	statusBody, _ := json.Marshal(statusReq)
	req = httptest.NewRequest(http.MethodPatch, "/v1/orders/"+strconv.FormatInt(order.ID, 10)+"/status", bytes.NewReader(statusBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid status update, got %d", rec.Code)
	}
}
