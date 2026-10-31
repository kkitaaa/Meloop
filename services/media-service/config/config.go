package config

import (
	"os"
	"strings"
)

// Config almacena la configuración para el microservicio media-service
type Config struct {
	Port                string
	DatabaseURL         string
	MinIOEndpoint       string
	MinIOPublicEndpoint string
	MinIOAccessKey      string
	MinIOSecretKey      string
	BucketName          string
	UseSSL              bool
	Region              string
}

// Load carga la configuración desde variables de entorno o valores por defecto
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/meloop?sslmode=disable"
	}

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}
	// Normalizar endpoint removiendo esquema si está presente para el SDK de MinIO
	minioEndpoint = strings.TrimPrefix(minioEndpoint, "http://")
	minioEndpoint = strings.TrimPrefix(minioEndpoint, "https://")

	minioPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if minioPublicEndpoint == "" {
		minioPublicEndpoint = "http://localhost:9000"
	}

	accessKey := os.Getenv("MINIO_MEDIA_USER")
	if accessKey == "" {
		accessKey = os.Getenv("MINIO_ROOT_USER")
	}
	if accessKey == "" {
		accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("MINIO_MEDIA_PASSWORD")
	if secretKey == "" {
		secretKey = os.Getenv("MINIO_ROOT_PASSWORD")
	}
	if secretKey == "" {
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	bucketName := os.Getenv("MINIO_MEDIA_BUCKET")
	if bucketName == "" {
		bucketName = os.Getenv("AWS_BUCKET_NAME")
	}
	if bucketName == "" {
		bucketName = "meloop-media"
	}

	useSSL := strings.ToLower(os.Getenv("MINIO_USE_SSL")) == "true" || strings.ToLower(os.Getenv("USE_SSL")) == "true"

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("MINIO_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}

	return &Config{
		Port:                port,
		DatabaseURL:         dbURL,
		MinIOEndpoint:       minioEndpoint,
		MinIOPublicEndpoint: minioPublicEndpoint,
		MinIOAccessKey:      accessKey,
		MinIOSecretKey:      secretKey,
		BucketName:          bucketName,
		UseSSL:              useSSL,
		Region:              region,
	}
}
