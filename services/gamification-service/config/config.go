package config

import "os"

type Config struct {
	DatabaseURL string
	RabbitMQURL string
	Port        string
}

func Load() *Config {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:15422/postgres?sslmode=disable"
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	port := os.Getenv("GAMIFICATION_SERVICE_PORT")
	if port == "" {
		port = "8089"
	}

	return &Config{
		DatabaseURL: databaseURL,
		RabbitMQURL: rabbitMQURL,
		Port:        port,
	}
}
