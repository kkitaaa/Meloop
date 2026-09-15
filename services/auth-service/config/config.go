package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL       string
	PasswordMinLength int
	RedisURL          string
	SessionTTL        time.Duration
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

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisPassword := os.Getenv("REDIS_PASSWORD")
		if redisPassword != "" {
			redisURL = "redis://:" + redisPassword + "@localhost:6379"
		} else {
			redisURL = "redis://localhost:6379"
		}
	}

	sessionTTL := 24 * time.Hour
	if ttlStr := os.Getenv("SESSION_TTL_HOURS"); ttlStr != "" {
		if val, err := strconv.Atoi(ttlStr); err == nil && val > 0 {
			sessionTTL = time.Duration(val) * time.Hour
		}
	}

	return &Config{
		DatabaseURL:       dbURL,
		PasswordMinLength: minLen,
		RedisURL:          redisURL,
		SessionTTL:        sessionTTL,
	}
}
