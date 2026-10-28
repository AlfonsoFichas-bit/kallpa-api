package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"

	"kallpa-api/internal/customers"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
)

func setupStoreData(t *testing.T) (*domain.Product, *domain.Customer) {
	testDB.Exec("DELETE FROM product_variants")
	testDB.Exec("DELETE FROM products")
	testDB.Exec("DELETE FROM customers")
	testDB.Exec("DELETE FROM telegram_identities")

	product := &domain.Product{
		StoreID:   1, // Match the StoreID of our auth flow tests if necessary or hardcode
		Name:      "Test Quinoa",
		BasePrice: decimal.NewFromInt(50),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Variants: []domain.ProductVariant{
			{
				SKU:             func(s string) *string { return &s }("SKU-1"),
				StockQuantity:   10,
				AdditionalPrice: decimal.NewFromInt(0),
				CreatedAt:       time.Now(),
			},
		},
	}
	if err := testDB.Create(product).Error; err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	telegramID := int64(112233)
	customer := &domain.Customer{
		StoreID:    1,
		TelegramID: &telegramID,
		FullName:   func(s string) *string { return &s }("Test Cust"),
		Status:     "lead",
		CreatedAt:  time.Now(),
		TelegramIdentity: &domain.TelegramIdentity{
			TelegramID: telegramID,
			FirstName:  func(s string) *string { return &s }("Test"),
			CreatedAt:  time.Now(),
		},
	}
	if err := testDB.Create(customer).Error; err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	return product, customer
}

func getAdminAuthToken(t *testing.T) string {
	// Re-use test setup from auth_test.go which ensures superadmin user is created
	setupTestUser(t, "admin_inv@example.com", "adminpass", domain.RoleSuperadmin)

	// Because setupTestUser uses store_id = 0/nil for admin by default. We must set it to 1 to match the product
	testDB.Exec("UPDATE users SET store_id = 1 WHERE email = 'admin_inv@example.com'")

	loginReq := map[string]string{"email": "admin_inv@example.com", "password": "adminpass"}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	testRouter.ServeHTTP(rec, req)

	var tokenRes map[string]string
	json.Unmarshal(rec.Body.Bytes(), &tokenRes)
	return tokenRes["access_token"]
}

func TestIntegration_InventoryFlow(t *testing.T) {
	product, _ := setupStoreData(t)
	token := getAdminAuthToken(t)

	// 1. List Products
	req := httptest.NewRequest(http.MethodGet, "/v1/products", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for products list, got %d. Body: %s", rec.Code, rec.Body.String())
	}
	var products []domain.Product
	json.Unmarshal(rec.Body.Bytes(), &products)
	if len(products) != 1 || products[0].Name != "Test Quinoa" {
		t.Errorf("expected to find test product, got %v", products)
	}

	// 2. Stock Movement (Success)
	variantID := product.Variants[0].ID
	stockReq := inventory.StockMovementRequest{
		VariantID: variantID,
		Delta:     -2,
	}
	body, _ := json.Marshal(stockReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/stock-movements", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for successful stock movement, got %d", rec.Code)
	}

	// 3. Stock Movement (Negative Stock rejection)
	stockReq.Delta = -20
	body, _ = json.Marshal(stockReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/stock-movements", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict for negative stock, got %d", rec.Code)
	}

	// Verify DB stock is 8 (10 - 2)
	var finalVariant domain.ProductVariant
	testDB.First(&finalVariant, variantID)
	if finalVariant.StockQuantity != 8 {
		t.Errorf("expected final stock to be 8, got %d", finalVariant.StockQuantity)
	}
}

func TestIntegration_CustomerFlow(t *testing.T) {
	_, _ = setupStoreData(t)
	token := getAdminAuthToken(t)

	// 1. List Customers
	req := httptest.NewRequest(http.MethodGet, "/v1/customers", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for customers list, got %d. Body: %s", rec.Code, rec.Body.String())
	}
	var custs []domain.Customer
	json.Unmarshal(rec.Body.Bytes(), &custs)
	if len(custs) != 1 {
		t.Errorf("expected to find 1 customer, got %d", len(custs))
	}

	// 2. Upsert Customer (Existing)
	upsertReq := customers.CustomerUpsertRequest{
		TelegramID: 112233,
		Name:       "Test Cust Updated",
	}
	body, _ := json.Marshal(upsertReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/customers", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for customer upsert, got %d", rec.Code)
	}

	// Verify DB count did not increase (idempotency)
	var count int64
	testDB.Model(&domain.Customer{}).Count(&count)
	if count != 1 {
		t.Errorf("expected customer count to be 1, got %d", count)
	}

	// Verify name updated
	var updatedCust domain.Customer
	testDB.First(&updatedCust)
	if *updatedCust.FullName != "Test Cust Updated" {
		t.Errorf("expected name to be updated")
	}

	// 3. Upsert Customer (New)
	upsertReq.TelegramID = 99999
	upsertReq.Name = "New Customer"
	body, _ = json.Marshal(upsertReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/customers", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	testDB.Model(&domain.Customer{}).Count(&count)
	if count != 2 {
		t.Errorf("expected customer count to be 2, got %d", count)
	}
}
