package config

import "os"

type Config struct {
	DatabaseURL          string
	RabbitMQURL          string
	RedisURL             string
	InternalServiceToken string
	Port                 string
	AuthServiceURL       string
}

func Load() *Config {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/meloop?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	internalServiceToken := os.Getenv("SOCIAL_SERVICE_INTERNAL_TOKEN")

	port := os.Getenv("SOCIAL_SERVICE_PORT")
	if port == "" {
		port = "8086"
	}

	return &Config{
		DatabaseURL:          databaseURL,
		RedisURL:             redisURL,
		RabbitMQURL:          rabbitMQURL,
		InternalServiceToken: internalServiceToken,
		Port:                 port,
		AuthServiceURL: authServiceURL,
}
