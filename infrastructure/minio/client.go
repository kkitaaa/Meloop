package minio

import (
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitClient() (*minio.Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:9000"
	}
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	if accessKeyID == "" {
		accessKeyID = "meloop"
	}
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	if secretAccessKey == "" {
		secretAccessKey = "Meloop.67"
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	log.Println("Cliente MinIO conectado exitosamente")
	return client, nil
}