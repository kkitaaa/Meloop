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
	GeneratePresignedDownloadURL(ctx context.Context, objectKey string, expires time.Duration) (string, error)
	ObjectExists(ctx context.Context, objectKey string) (bool, int64, string, error)
	DeleteObject(ctx context.Context, objectKey string) error
}

type minioStorageRepository struct {
	client        *minio.Client
	signingClient *minio.Client
	bucketName    string
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
		u, err := url.Parse(publicEndpoint)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") ||
			(u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("MINIO_PUBLIC_ENDPOINT debe ser una URL HTTP(S) absoluta")
		}
		publicHost = u.Host
		publicSSL = u.Scheme == "https"
		sClient, err := minio.New(publicHost, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: publicSSL,
			Region: region,
		})
		if err != nil {
			return nil, fmt.Errorf("error al inicializar cliente de firma pública MinIO: %w", err)
		}
		signingClient = sClient
	}

	return &minioStorageRepository{
		client:        minioClient,
		signingClient: signingClient,
		bucketName:    bucketName,
	}, nil
}

// NewMinIOStorageRepositoryWithClient permite inyectar un cliente MinIO existente (útil para pruebas)
func NewMinIOStorageRepositoryWithClient(client *minio.Client, bucketName, _ string) StorageRepository {
	return &minioStorageRepository{
		client:        client,
		signingClient: client,
		bucketName:    bucketName,
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

// GeneratePresignedDownloadURL genera una URL prefirmada GET para lectura temporal.
func (r *minioStorageRepository) GeneratePresignedDownloadURL(ctx context.Context, objectKey string, expires time.Duration) (string, error) {
	clientToUse := r.signingClient
	if clientToUse == nil {
		clientToUse = r.client
	}
	if clientToUse == nil {
		return "", errors.New("cliente de almacenamiento no inicializado")
	}

	u, err := clientToUse.PresignedGetObject(ctx, r.bucketName, objectKey, expires, nil)
	if err != nil {
		return "", fmt.Errorf("error al generar URL prefirmada de descarga: %w", err)
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
