package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/minio/minio-go/v7"
)

const (
	bucketName   = "meloop-media"
	maxImageSize = 5 << 20  // 5 MB
	maxAudioSize = 15 << 20 // 15 MB
)

// UploadMedia contiene la lógica para recibir, validar (tamaño/formato) y procesar imágenes y audios.
func UploadMedia(minioClient *minio.Client, file *multipart.FileHeader) (string, error) {
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

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	objectName := file.Filename
	_, err = minioClient.PutObject(context.Background(), bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("error subiendo archivo a MinIO: %w", err)
	}

	endpoint := minioClient.EndpointURL().Host
	publicURL := fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, objectName)

	return publicURL, nil
}

// DeleteMedia elimina de MinIO el objeto asociado a una URL multimedia
func DeleteMedia(ctx context.Context, minioClient *minio.Client, mediaURL string) error {
	if minioClient == nil || strings.TrimSpace(mediaURL) == "" {
		return nil
	}

	objectName := strings.TrimSpace(mediaURL)
	prefix := "/" + bucketName + "/"
	if idx := strings.Index(objectName, prefix); idx != -1 {
		objectName = objectName[idx+len(prefix):]
	} else if strings.HasPrefix(objectName, bucketName+"/") {
		objectName = strings.TrimPrefix(objectName, bucketName+"/")
	}

	if qIdx := strings.Index(objectName, "?"); qIdx != -1 {
		objectName = objectName[:qIdx]
	}

	if objectName == "" {
		return nil
	}

	if err := minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("error eliminando objeto %s de MinIO: %w", objectName, err)
	}

	return nil
}
