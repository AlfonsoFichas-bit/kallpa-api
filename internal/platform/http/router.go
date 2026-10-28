package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"gorm.io/gorm"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/config"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

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
				if he, ok := err.(*echo.HTTPError); ok {
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
