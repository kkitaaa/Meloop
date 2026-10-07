package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
)

const (
	bucketName    = "meloop-media"
	maxImageSize  = 5 << 20  // 5 MB
	maxAudioSize  = 15 << 20 // 15 MB
)

// UploadMedia contiene la lógica para recibir, validar (tamaño/formato) y procesar imágenes y audios[cite: 9].
func UploadMedia(minioClient *minio.Client, file *multipart.FileHeader) (string, error) {
	// 1. Validar el formato y tamaño
	contentType := file.Header.Get("Content-Type")
	
	switch contentType {
	case "image/jpeg", "image/png":
		if file.Size > maxImageSize {
			return "", errors.New("la imagen supera el tamaño máximo de 5MB")
		}
	case "audio/mpeg", "audio/wav":
		if file.Size > maxAudioSize {
			return "", errors.New("el audio supera el tamaño máximo de 15MB")
		}
	default:
		return "", fmt.Errorf("formato no soportado: %s", contentType)
	}

	// 2. Abrir el archivo
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 3. Subir el archivo al servidor multimedia[cite: 9].
	objectName := file.Filename // En producción, deberías generar un UUID para evitar colisiones
	_, err = minioClient.PutObject(context.Background(), bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("error subiendo archivo a MinIO: %w", err)
	}

	// 4. Retornar las URLs públicas generadas para luego guardarlas en la base de datos[cite: 9].
	endpoint := minioClient.EndpointURL().Host
	publicURL := fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, objectName)

	return publicURL, nil
}