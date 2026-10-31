package repositories

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// StorageRepository define las operaciones sobre el almacenamiento de objetos MinIO/S3
type StorageRepository interface {
	GeneratePresignedUploadURL(ctx context.Context, objectKey string, contentType string, expires time.Duration) (string, error)
	ObjectExists(ctx context.Context, objectKey string) (bool, int64, string, error)
	DeleteObject(ctx context.Context, objectKey string) error
}

type minioStorageRepository struct {
	client         *minio.Client
	signingClient  *minio.Client
	bucketName     string
	publicEndpoint string
}

// NewMinIOStorageRepository crea una nueva instancia de StorageRepository conectada a MinIO/S3
func NewMinIOStorageRepository(endpoint, accessKey, secretKey, bucketName string, useSSL bool, region, publicEndpoint string) (StorageRepository, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error al inicializar cliente MinIO: %w", err)
	}

	signingClient := minioClient
	if publicEndpoint != "" {
		publicHost := publicEndpoint
		publicSSL := useSSL
		if u, err := url.Parse(publicEndpoint); err == nil && u.Host != "" {
			publicHost = u.Host
			if u.Scheme == "https" {
				publicSSL = true
			} else if u.Scheme == "http" {
				publicSSL = false
			}
		}
		if sClient, err := minio.New(publicHost, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: publicSSL,
			Region: region,
		}); err == nil {
			signingClient = sClient
		}
	}

	return &minioStorageRepository{
		client:         minioClient,
		signingClient:  signingClient,
		bucketName:     bucketName,
		publicEndpoint: publicEndpoint,
	}, nil
}

// NewMinIOStorageRepositoryWithClient permite inyectar un cliente MinIO existente (útil para pruebas)
func NewMinIOStorageRepositoryWithClient(client *minio.Client, bucketName, publicEndpoint string) StorageRepository {
	return &minioStorageRepository{
		client:         client,
		signingClient:  client,
		bucketName:     bucketName,
		publicEndpoint: publicEndpoint,
	}
}

// GeneratePresignedUploadURL genera una URL prefirmada PUT para subida directa
func (r *minioStorageRepository) GeneratePresignedUploadURL(ctx context.Context, objectKey string, contentType string, expires time.Duration) (string, error) {
	clientToUse := r.client
	if r.signingClient != nil {
		clientToUse = r.signingClient
	}
	if clientToUse == nil {
		return "", errors.New("cliente de almacenamiento no inicializado")
	}

	u, err := clientToUse.PresignedPutObject(ctx, r.bucketName, objectKey, expires)
	if err != nil {
		return "", fmt.Errorf("error al generar URL prefirmada de subida: %w", err)
	}

	return u.String(), nil
}

// ObjectExists verifica la presencia física del objeto en MinIO/S3
func (r *minioStorageRepository) ObjectExists(ctx context.Context, objectKey string) (bool, int64, string, error) {
	if r.client == nil {
		return false, 0, "", errors.New("cliente de almacenamiento no inicializado")
	}

	info, err := r.client.StatObject(ctx, r.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.Code == "ResourceNotFound" || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return false, 0, "", nil
		}
		return false, 0, "", fmt.Errorf("error al consultar objeto en MinIO/S3: %w", err)
	}

	return true, info.Size, info.ContentType, nil
}

// DeleteObject elimina un objeto del bucket
func (r *minioStorageRepository) DeleteObject(ctx context.Context, objectKey string) error {
	if r.client == nil {
		return errors.New("cliente de almacenamiento no inicializado")
	}

	return r.client.RemoveObject(ctx, r.bucketName, objectKey, minio.RemoveObjectOptions{})
}

// adjustPublicURL reemplaza el host interno por el endpoint público accesible para el cliente
func (r *minioStorageRepository) adjustPublicURL(rawURL string) string {
	if r.publicEndpoint == "" {
		return rawURL
	}

	parsedPublic, err := url.Parse(r.publicEndpoint)
	if err != nil || parsedPublic.Host == "" {
		return rawURL
	}

	parsedGenerated, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsedGenerated.Scheme = parsedPublic.Scheme
	parsedGenerated.Host = parsedPublic.Host

	return parsedGenerated.String()
}
