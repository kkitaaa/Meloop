package config

import "os"

// Config almacena la configuración requerida para el microservicio de usuarios
type Config struct {
	DatabaseURL    string
	AuthServiceURL string
	Port           string
}

// Load carga las variables de entorno o utiliza valores por defecto locales
func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@localhost:15422/postgres?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8083"
	}

	return &Config{
		DatabaseURL:    dbURL,
		AuthServiceURL: authServiceURL,
		Port:           port,
	}
}
