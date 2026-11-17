package config

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/meloop/services/common/logging"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient crea y verifica la conexión a Redis desde variables de entorno
func NewRedisClient(logger *logging.Logger) *redis.Client {
	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	db := 0
	if dbStr := strings.TrimSpace(os.Getenv("REDIS_DB")); dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil && parsed >= 0 {
			db = parsed
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		if logger != nil {
			logger.Warn("redis_unavailable_on_startup", "addr", addr, "error", err.Error())
		}
	} else if logger != nil {
		logger.Info("redis_connected", "addr", addr, "db", db)
	}

	return client
}
