package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/config"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/platform/db"
	kallpahttp "kallpa-api/internal/platform/http"
)

var (
	testDB     *gorm.DB
	testRouter *echo.Echo
	cfg        *config.Config
)

func TestMain(m *testing.M) {
	cfg = &config.Config{
		DBName:    "file::memory:?cache=shared",
		Port:      "8080",
		LogLevel:  "error",
		JWTSecret: "testsecret",
	}

	database, err := db.Init(cfg.DBName)
	if err != nil {
		os.Exit(1)
	}
	testDB = database

	_ = testDB.AutoMigrate(
		&domain.User{},
		&domain.Product{},
		&domain.ProductVariant{},
		&domain.TelegramIdentity{},
		&domain.Customer{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.DeliveryZone{},
	)

	testRouter = kallpahttp.NewRouter(testDB, cfg)

	// Add a dummy protected route for testing
	authService := auth.NewService(auth.NewUserRepository(testDB), cfg.JWTSecret)
	protectedGroup := testRouter.Group("/v1/protected")
	protectedGroup.Use(auth.AuthMiddleware(authService))
	protectedGroup.GET("/admin-only", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}, auth.RequireRole(domain.RoleSuperadmin))

	code := m.Run()

	os.Exit(code)
}

func setupTestUser(t *testing.T, email, password string, role domain.UserRole) *domain.User {
	testDB.Exec("DELETE FROM users") // clean up
	authService := auth.NewService(auth.NewUserRepository(testDB), cfg.JWTSecret)
	hash, _ := authService.HashPassword(password)

	user := &domain.User{
		Name:         "Test User",
		Email:        email,
		PasswordHash: hash,
		Role:         role,
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestIntegration_AuthFlow(t *testing.T) {
	email := "agent@example.com"
	password := "agentpass"
	setupTestUser(t, email, password, domain.RoleSalesAgent)

	// 1. Test Login Success
	loginReq := auth.LoginRequest{
		Email:    email,
		Password: password,
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", rec.Code)
	}

	var tokenRes auth.TokenResponse
	json.Unmarshal(rec.Body.Bytes(), &tokenRes)
	if tokenRes.AccessToken == "" || tokenRes.RefreshToken == "" {
		t.Fatalf("expected both tokens in response")
	}

	// 2. Test Invalid Credentials
	loginReq.Password = "wrong"
	body, _ = json.Marshal(loginReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong credentials, got %d", rec.Code)
	}

	// 3. Test Refresh Success
	refreshReq := auth.RefreshRequest{
		RefreshToken: tokenRes.RefreshToken,
	}
	body, _ = json.Marshal(refreshReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for refresh, got %d", rec.Code)
	}
	var newTokenRes auth.TokenResponse
	json.Unmarshal(rec.Body.Bytes(), &newTokenRes)
	if newTokenRes.AccessToken == "" {
		t.Errorf("expected new access token")
	}

	// 4. Test Protected Route - Missing Token
	req = httptest.NewRequest(http.MethodGet, "/v1/protected/admin-only", nil)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing token on protected route, got %d", rec.Code)
	}

	// 5. Test Protected Route - Insufficient Role
	req = httptest.NewRequest(http.MethodGet, "/v1/protected/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken) // SalesAgent token
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for insufficient role, got %d", rec.Code)
	}

	// 6. Test Protected Route - Success (Superadmin)
	setupTestUser(t, "admin@example.com", "adminpass", domain.RoleSuperadmin)
	adminLoginReq := auth.LoginRequest{Email: "admin@example.com", Password: "adminpass"}
	body, _ = json.Marshal(adminLoginReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	var adminTokenRes auth.TokenResponse
	json.Unmarshal(rec.Body.Bytes(), &adminTokenRes)

	req = httptest.NewRequest(http.MethodGet, "/v1/protected/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+adminTokenRes.AccessToken)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for superadmin on protected route, got %d", rec.Code)
	}
}
