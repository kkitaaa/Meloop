package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	bucketName   = "meloop-media"
	maxImageSize = 5 << 20  // 5 MB
	maxAudioSize = 15 << 20 // 15 MB
)

// MediaStorage abstrae el almacenamiento de archivos multimedia para permitir compensación y tests
type MediaStorage interface {
	Upload(ctx context.Context, file *multipart.FileHeader) (string, error)
	Delete(ctx context.Context, mediaURL string) error
}

// MinioMediaStorage implementa MediaStorage sobre MinIO
type MinioMediaStorage struct {
	client *minio.Client
}

func NewMinioMediaStorage(client *minio.Client) *MinioMediaStorage {
	return &MinioMediaStorage{client: client}
}

func (m *MinioMediaStorage) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	return UploadMediaWithContext(ctx, m.client, file)
}

func (m *MinioMediaStorage) Delete(ctx context.Context, mediaURL string) error {
	return DeleteMedia(ctx, m.client, mediaURL)
}

// UploadMedia mantiene compatibilidad con el endpoint /api/v1/posts/media
func UploadMedia(minioClient *minio.Client, file *multipart.FileHeader) (string, error) {
	return UploadMediaWithContext(context.Background(), minioClient, file)
}

// UploadMediaWithContext valida formato/tamaño y sube el archivo a MinIO con nombre único UUID
func UploadMediaWithContext(ctx context.Context, minioClient *minio.Client, file *multipart.FileHeader) (string, error) {
	if minioClient == nil {
		return "", errors.New("cliente de MinIO no disponible")
	}
	if file == nil {
		return "", errors.New("archivo multimedia requerido")
	}

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

	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%s%s", uuid.NewString(), ext)

	_, err = minioClient.PutObject(ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("error subiendo archivo a MinIO: %w", err)
	}

	endpoint := minioClient.EndpointURL().Host
	publicURL := fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, objectName)

	return publicURL, nil
}

// DeleteMedia elimina de MinIO un objeto previamente subido (compensación ante rollback o borrado de post)
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
