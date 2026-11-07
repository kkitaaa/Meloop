package config

import "os"

// Config almacena la configuración requerida para el microservicio de usuarios
type Config struct {
	DatabaseURL     string
	Port            string
	AuthServiceURL  string
	MediaServiceURL string
}

// Load carga las variables de entorno o utiliza valores por defecto locales
func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/meloop?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8083"
	}

	mediaServiceURL := os.Getenv("MEDIA_SERVICE_URL")
	if mediaServiceURL == "" {
		mediaServiceURL = "http://localhost:8085"
	}

	return &Config{
		DatabaseURL:     dbURL,
		Port:            port,
		AuthServiceURL:  authServiceURL,
		MediaServiceURL: mediaServiceURL,
	}
}
