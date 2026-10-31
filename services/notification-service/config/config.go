package config

import "os"

type Config struct {
	DatabaseURL string
	RabbitMQURL string
}

func Load() Config {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/meloop?sslmode=disable"
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	return Config{DatabaseURL: databaseURL, RabbitMQURL: rabbitMQURL}
}
