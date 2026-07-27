package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   []byte
	JWTExpiry   time.Duration
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		JWTSecret:   []byte(jwtSecret),
		JWTExpiry:   72 * time.Hour,
	}, nil
}
