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
	RecoveryTokenTTL  time.Duration
	RecoveryURLBase   string
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPassword      string
	SMTPFrom          string
}

func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@localhost:15422/postgres?sslmode=disable"
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

	recoveryTTL := 30 * time.Minute
	if rTTLStr := os.Getenv("RECOVERY_TOKEN_TTL_MINUTES"); rTTLStr != "" {
		if val, err := strconv.Atoi(rTTLStr); err == nil && val > 0 {
			recoveryTTL = time.Duration(val) * time.Minute
		}
	}

	recoveryURLBase := os.Getenv("RECOVERY_URL_BASE")
	if recoveryURLBase == "" {
		recoveryURLBase = "http://localhost:8080/auth/password-recovery/reset?token="
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 587
	if portStr := os.Getenv("SMTP_PORT"); portStr != "" {
		if val, err := strconv.Atoi(portStr); err == nil && val > 0 {
			smtpPort = val
		}
	}
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")
	if smtpFrom == "" {
		smtpFrom = "noreply@meloop.app"
	}

	return &Config{
		DatabaseURL:       dbURL,
		PasswordMinLength: minLen,
		RedisURL:          redisURL,
		SessionTTL:        sessionTTL,
		RecoveryTokenTTL:  recoveryTTL,
		RecoveryURLBase:   recoveryURLBase,
		SMTPHost:          smtpHost,
		SMTPPort:          smtpPort,
		SMTPUser:          smtpUser,
		SMTPPassword:      smtpPass,
		SMTPFrom:          smtpFrom,
	}
}
