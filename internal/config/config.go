package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config represents the application configuration settings.
type Config struct {
	DBName       string
	Port         string
	LogLevel     string
	JWTSecret    string
	ServiceToken string
}

// Load retrieves configuration from environment variables and an optional .env file.
func Load() *Config {
	// Try to load .env file, ignore error if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		DBName:       getEnv("DB_NAME", "kallpa.db"),
		Port:         getEnv("PORT", "8080"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		JWTSecret:    getEnv("JWT_SECRET", ""),
		ServiceToken: getEnv("SERVICE_TOKEN", ""),
	}

	if cfg.JWTSecret == "" {
		log.Println("WARNING: JWT_SECRET is empty! Using default for development only.")
		cfg.JWTSecret = "secret"
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
