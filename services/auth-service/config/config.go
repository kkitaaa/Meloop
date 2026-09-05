package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	PasswordMinLength int
}

func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/meloop?sslmode=disable"
	}

	minLen := 8
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
