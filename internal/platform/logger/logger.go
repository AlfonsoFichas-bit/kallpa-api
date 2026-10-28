package logger

import (
	"log/slog"
	"os"
)

// Init creates and sets the default global slog logger.
func Init(level string) *slog.Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: l,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
