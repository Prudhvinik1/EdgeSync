package config

import (
	"os"
	"time"
	"errors"
)


type Config struct {
	ServerPort string
	DatabaseURL string
	RedisURL string
	JWTSecret string
	JWTExpiry time.Duration
}

func LoadConfig() (*Config, error) {
	expiryStr := getEnv("JWT_EXPIRY", "24h")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		return nil, errors.New("invalid JWT_EXPIRY format")
	}

	cfg := &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTExpiry:   expiry,
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return nil, errors.New("REDIS_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	return cfg, nil
}

// Helper: get env with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}