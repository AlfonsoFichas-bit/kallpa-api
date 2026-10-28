package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *echo.Echo {
	e := echo.New()

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
			// c.Response() in Echo v5 returns an interface that embeds http.ResponseWriter
			// We can get the actual *echo.Response using type assertion or by directly accessing Response
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
