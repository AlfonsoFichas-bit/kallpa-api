package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
)

func main() {
	e := echo.New()
	e.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		e.Logger.Error("server failed to start", "error", err)
	}
}
