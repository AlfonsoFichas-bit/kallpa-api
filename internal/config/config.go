package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBName    string
	Port      string
	LogLevel  string
	JWTSecret string
}

func Load() *Config {
	// Try to load .env file, ignore error if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		DBName:    getEnv("DB_NAME", "kallpa.db"),
		Port:      getEnv("PORT", "8080"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		JWTSecret: getEnv("JWT_SECRET", ""),
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
