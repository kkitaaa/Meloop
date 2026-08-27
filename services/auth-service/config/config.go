package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	PasswordMinLength int
}

// Load loads the configuration from environment variables
func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default fallback for local development
		dbURL = "postgres://postgres:postgres@localhost:5432/meloop?sslmode=disable"
	}

	minLen := 8 // Baseline default length as implementation decision
	if minLenStr := os.Getenv("PASSWORD_MIN_LENGTH"); minLenStr != "" {
		if val, err := strconv.Atoi(minLenStr); err == nil && val > 0 {
			minLen = val
		}
	}

	return &Config{
		DatabaseURL:       dbURL,
		PasswordMinLength: minLen,
	}
}
