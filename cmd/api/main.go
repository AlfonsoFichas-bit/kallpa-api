package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kallpa-api/internal/config"
	kallpadomain "kallpa-api/internal/domain"
	"kallpa-api/internal/platform/db"
	kallpahttp "kallpa-api/internal/platform/http"
	"kallpa-api/internal/platform/logger"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Initialize logger
	logg := logger.Init(cfg.LogLevel)
	slog.SetDefault(logg)
	slog.Info("Starting Kallpa Go API", "port", cfg.Port)

	// 3. Initialize database
	database, err := db.Init(cfg.DBName)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	slog.Info("Database connected", "db_name", cfg.DBName)

	if err := database.AutoMigrate(&kallpadomain.User{}); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	// 4. Initialize HTTP router
	router := kallpahttp.NewRouter(database, cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// 5. Start Server gracefully
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server stopped", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting")
}
