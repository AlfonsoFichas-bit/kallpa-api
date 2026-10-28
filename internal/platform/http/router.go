package http

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"gorm.io/gorm"

	"kallpa-api/internal/ai"
	"kallpa-api/internal/auth"
	"kallpa-api/internal/chat"
	"kallpa-api/internal/config"
	"kallpa-api/internal/customers"
	"kallpa-api/internal/inventory"
	"kallpa-api/internal/notifications"
	"kallpa-api/internal/sales"
	"kallpa-api/internal/shipping"
	"kallpa-api/internal/stores"
)

// CustomValidator wraps the standard validator for Echo.
type CustomValidator struct {
	validator *validator.Validate
}

// Validate executes the validation rules defined in the structs.
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// NewRouter initializes and configures the Echo web framework.
func NewRouter(db *gorm.DB, cfg *config.Config) *echo.Echo {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middlewares
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())

	corsConfig := middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}
	e.Use(middleware.CORSWithConfig(corsConfig))

	e.Use(slogMiddleware())

	// Basic routes
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/ready", func(c *echo.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "error", "message": "database unavailable"})
		}
		if err := sqlDB.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "error", "message": "database not responding"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	// Auth module
	userRepo := auth.NewUserRepository(db)
	authService := auth.NewService(userRepo, cfg.JWTSecret)
	authHandler := auth.NewAuthHandler(authService)

	v1 := e.Group("/v1")
	authGroup := v1.Group("/auth")
	auth.RegisterHandlers(authGroup, authHandler)

	// Protected routes
	protected := v1.Group("")
	protected.Use(auth.AuthMiddleware(authService))

	// Inventory module
	inventoryRepo := inventory.NewProductRepository(db)
	inventoryService := inventory.NewService(inventoryRepo)
	inventoryHandler := inventory.NewInventoryHandler(inventoryService)
	inventory.RegisterHandlers(protected, inventoryHandler)

	// Customers module
	customerRepo := customers.NewCustomerRepository(db)
	customerService := customers.NewService(customerRepo)
	customerHandler := customers.NewCustomerHandler(customerService)
	customers.RegisterHandlers(protected, customerHandler)

	// Sales module
	salesRepo := sales.NewOrderRepository(db)
	salesService := sales.NewService(salesRepo, inventoryRepo, customerRepo)
	salesHandler := sales.NewSalesHandler(salesService)
	sales.RegisterHandlers(protected, salesHandler)

	// Shipping module
	shippingRepo := shipping.NewDeliveryZoneRepository(db)
	shippingService := shipping.NewService(shippingRepo, salesService)
	shippingHandler := shipping.NewShippingHandler(shippingService)
	shipping.RegisterHandlers(protected, shippingHandler)

	// AI and Chat Module
	aiProvider := ai.NewDummyProvider()
	chatRepo := chat.NewRepository(db)
	chatService := chat.NewService(chatRepo, customerService, aiProvider)
	chatHandler := chat.NewHandler(chatService)
	chat.RegisterHandlers(protected, chatHandler)

	// Notifications module
	notificationRepo := notifications.NewRepository(db)
	notificationService := notifications.NewService(notificationRepo)
	notificationHandler := notifications.NewHandler(notificationService)
	notifications.RegisterHandlers(protected, notificationHandler)

	// Stores module
	storeRepo := stores.NewRepository(db)
	storeService := stores.NewService(storeRepo)
	storeHandler := stores.NewHandler(storeService)
	stores.RegisterHandlers(protected, storeHandler)

	return e
}

func slogMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			req := c.Request()
			reqID := req.Header.Get(echo.HeaderXRequestID)

			err := next(c)

			// Get status code
			var status int
			if eRes, ok := c.Response().(*echo.Response); ok {
				status = eRes.Status
			}

			if err != nil {
				var he *echo.HTTPError
				if errors.As(err, &he) {
					status = he.Code
				} else if status == 0 {
					status = http.StatusInternalServerError
				}
			}

			slog.Info("Request",
				"method", req.Method,
				"uri", req.RequestURI,
				"status", status,
				"latency", time.Since(start).String(),
				"request_id", reqID,
			)

			return err
		}
	}
}
