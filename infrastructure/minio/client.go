package minio

import (
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// InitClient configura el cliente de MinIO en Go para establecer una conexión segura.
func InitClient() (*minio.Client, error) {
	// Usamos 127.0.0.1 por defecto para evitar problemas de IPv6 en Windows
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:9000" 
	}
	
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	if accessKeyID == "" {
		accessKeyID = "meloop" // Tu usuario local
	}
	
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	if secretAccessKey == "" {
		secretAccessKey = "Meloop.67" // Tu contraseña local
	}
	
	useSSL := false // Falso para desarrollo local

	// Inicializamos el cliente usando las credenciales
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Cliente MinIO conectado exitosamente")
	return client, nil
}