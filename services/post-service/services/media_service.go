package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	bucketName   = "meloop-media"
	maxImageSize = 5 << 20  // 5 MB
	maxAudioSize = 15 << 20 // 15 MB
)

var (
	ErrUnauthorizedMediaAccess = errors.New("el recurso multimedia no pertenece al usuario autenticado")
	validExtRegex              = regexp.MustCompile(`^\.[a-zA-Z0-9]{1,10}$`)
)

// MediaStorage abstrae el almacenamiento de archivos multimedia con aislamiento por propietario
type MediaStorage interface {
	Upload(ctx context.Context, ownerID string, file *multipart.FileHeader) (string, error)
	Delete(ctx context.Context, ownerID string, mediaURL string) error
	EndpointHost() string
}

// MinioMediaStorage implementa MediaStorage sobre MinIO
type MinioMediaStorage struct {
	client *minio.Client
}

func NewMinioMediaStorage(client *minio.Client) *MinioMediaStorage {
	return &MinioMediaStorage{client: client}
}

func (m *MinioMediaStorage) Upload(ctx context.Context, ownerID string, file *multipart.FileHeader) (string, error) {
	_, publicURL, err := UploadMediaForOwner(ctx, m.client, ownerID, file)
	return publicURL, err
}

func (m *MinioMediaStorage) Delete(ctx context.Context, ownerID string, mediaURL string) error {
	return DeleteMediaForOwner(ctx, m.client, ownerID, mediaURL)
}

func (m *MinioMediaStorage) EndpointHost() string {
	if m.client == nil || m.client.EndpointURL() == nil {
		return ""
	}
	return m.client.EndpointURL().Host
}

func validateOwnerID(ownerID string) (string, error) {
	trimmed := strings.TrimSpace(ownerID)
	if trimmed == "" {
		return "", errors.New("usuario autenticado requerido para gestionar multimedia")
	}
	if strings.Contains(trimmed, "/") ||
		strings.Contains(trimmed, "\\") ||
		strings.Contains(trimmed, "..") ||
		strings.ContainsAny(trimmed, " \t\r\n") {
		return "", errors.New("identificador de usuario inválido para almacenamiento multimedia")
	}
	return trimmed, nil
}

func buildOwnedObjectKey(ownerID, filename string) (string, error) {
	cleanOwner, err := validateOwnerID(ownerID)
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" && !validExtRegex.MatchString(ext) {
		ext = ""
	}
	return fmt.Sprintf("users/%s/%s%s", cleanOwner, uuid.NewString(), ext), nil
}

// ExtractOwnedMinioObjectKey valida que mediaURL o clave pertenezca al bucket interno y al expectedOwnerID
func ExtractOwnedMinioObjectKey(rawMedia string, expectedOwnerID string, minioHost string) (string, bool) {
	trimmed := strings.TrimSpace(rawMedia)
	if trimmed == "" {
		return "", false
	}

	cleanOwner, err := validateOwnerID(expectedOwnerID)
	if err != nil {
		return "", false
	}

	var candidateKey string
	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return "", false
		}
		cleanHost := strings.TrimSpace(minioHost)
		if cleanHost == "" || !strings.EqualFold(parsed.Host, cleanHost) {
			return "", false
		}
		escapedPath := strings.ToLower(parsed.EscapedPath())
		if strings.Contains(escapedPath, "%2f") || strings.Contains(escapedPath, "%2e") {
			return "", false
		}
		prefix := "/" + bucketName + "/"
		if !strings.HasPrefix(parsed.Path, prefix) {
			return "", false
		}
		candidateKey = strings.TrimPrefix(parsed.Path, prefix)
	} else {
		candidateKey = strings.TrimPrefix(trimmed, "/")
		if strings.HasPrefix(candidateKey, bucketName+"/") {
			candidateKey = strings.TrimPrefix(candidateKey, bucketName+"/")
		}
	}

	if strings.Contains(candidateKey, "..") || strings.Contains(candidateKey, "\\") {
		return "", false
	}

	parts := strings.Split(candidateKey, "/")
	if len(parts) != 3 || parts[0] != "users" || parts[1] != cleanOwner {
		return "", false
	}

	filename := parts[2]
	base := filename
	if dotIdx := strings.IndexByte(filename, '.'); dotIdx != -1 {
		base = filename[:dotIdx]
		ext := filename[dotIdx:]
		if !validExtRegex.MatchString(ext) {
			return "", false
		}
	}

	if _, err := uuid.Parse(base); err != nil {
		return "", false
	}

	return candidateKey, true
}

// UploadMediaForOwner valida formato/tamaño y sube el archivo a MinIO bajo users/<ownerID>/<uuid>.<ext>
func UploadMediaForOwner(ctx context.Context, minioClient *minio.Client, ownerID string, file *multipart.FileHeader) (string, string, error) {
	if minioClient == nil {
		return "", "", errors.New("cliente de MinIO no disponible")
	}
	if file == nil {
		return "", "", errors.New("archivo multimedia requerido")
	}

	objectName, err := buildOwnedObjectKey(ownerID, file.Filename)
	if err != nil {
		return "", "", err
	}

	contentType := file.Header.Get("Content-Type")
	switch contentType {
	case "image/jpeg", "image/png":
		if file.Size > maxImageSize {
			return "", "", errors.New("la imagen supera el tamaño máximo de 5MB")
		}
	case "audio/mpeg", "audio/wav":
		if file.Size > maxAudioSize {
			return "", "", errors.New("el audio supera el tamaño máximo de 15MB")
		}
	default:
		return "", "", fmt.Errorf("formato no soportado: %s", contentType)
	}

	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	_, err = minioClient.PutObject(ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", fmt.Errorf("error subiendo archivo a MinIO: %w", err)
	}

	endpoint := minioClient.EndpointURL().Host
	publicURL := fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, objectName)

	return objectName, publicURL, nil
}

// DeleteMediaForOwner elimina de MinIO únicamente objetos generados por el servidor y pertenecientes a ownerID
func DeleteMediaForOwner(ctx context.Context, minioClient *minio.Client, ownerID string, mediaURL string) error {
	if minioClient == nil || strings.TrimSpace(mediaURL) == "" {
		return nil
	}

	minioHost := ""
	if minioClient.EndpointURL() != nil {
		minioHost = minioClient.EndpointURL().Host
	}

	objectName, ok := ExtractOwnedMinioObjectKey(mediaURL, ownerID, minioHost)
	if !ok {
		// Enlaces externos o claves no pertenecientes al autor nunca provocan borrados en MinIO
		return nil
	}

	if err := minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("error eliminando objeto %s de MinIO: %w", objectName, err)
	}

	return nil
}
