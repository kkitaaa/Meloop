package config

import "os"

type Config struct {
	DatabaseURL    string
	RabbitMQURL    string
	AuthServiceURL string
	Port           string
}

func Load() Config {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres@localhost:15422/postgres?sslmode=disable"
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8083"
	}

	port := os.Getenv("GAMIFICATION_SERVICE_PORT")
	if port == "" {
		port = "8088"
	}

	return Config{DatabaseURL: databaseURL, RabbitMQURL: rabbitMQURL, AuthServiceURL: authServiceURL, Port: port}
}
