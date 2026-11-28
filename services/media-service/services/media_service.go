package services

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/meloop/media-service/models"
	"github.com/meloop/media-service/repositories"
)

var (
	// ErrObjectNotFound indica que el objeto no existe en el almacenamiento
	ErrObjectNotFound = errors.New("OBJECT_NOT_FOUND")
	// ErrStorageUnreachable indica falla de comunicación con el almacenamiento de objetos
	ErrStorageUnreachable = errors.New("STORAGE_UNREACHABLE")
	// ErrUnauthorized indica que no se proporcionó una sesión de usuario válida
	ErrUnauthorized = errors.New("UNAUTHORIZED")
)

// ValidationError representa un error de validación en los parámetros multimedia
type ValidationError struct {
	Field   string `json:"field"`
	Issue   string `json:"issue"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// MediaService define la interfaz de operaciones de negocio para multimedia de perfiles y publicaciones.
type MediaService interface {
	CreatePresignedUpload(ctx context.Context, userID string, req *models.PresignedUploadRequest) (*models.PresignedUploadResponse, error)
	CreatePresignedDownload(ctx context.Context, userID, objectKey string) (*models.PresignedDownloadResponse, error)
	VerifyMedia(ctx context.Context, objectKey string) (*models.VerifyMediaResponse, error)
}

type mediaService struct {
	storageRepo repositories.StorageRepository
}

// NewMediaService crea una nueva instancia del servicio multimedia
func NewMediaService(storageRepo repositories.StorageRepository) MediaService {
	return &mediaService{storageRepo: storageRepo}
}

// CreatePresignedUpload valida el archivo y genera la URL prefirmada de subida para perfiles o publicaciones.
func (s *mediaService) CreatePresignedUpload(ctx context.Context, userID string, req *models.PresignedUploadRequest) (*models.PresignedUploadResponse, error) {
	ownerID := strings.TrimSpace(userID)
	if ownerID == "" {
		return nil, ErrUnauthorized
	}

	if req == nil {
		return nil, &ValidationError{
			Field:   "request",
			Issue:   "required",
			Message: "Los datos de la solicitud multimedia son obligatorios",
		}
	}

	resourceType := req.GetResourceType()
	if resourceType == "" {
		return nil, &ValidationError{
			Field:   "resource_type",
			Issue:   "required",
			Message: "El tipo de recurso es obligatorio ('profile_photo', 'banner' o 'post_attachment')",
		}
	}

	if resourceType != models.ResourceTypeProfilePhoto &&
		resourceType != models.ResourceTypeBanner &&
		resourceType != models.ResourceTypePostAttachment {
		return nil, &ValidationError{
			Field:   "resource_type",
			Issue:   "unsupported_type",
			Message: fmt.Sprintf("Tipo de recurso no soportado para esta operación: '%s'.", resourceType),
		}
	}

	contentType := req.GetContentType()
	if contentType == "" {
		return nil, &ValidationError{
			Field:   "content_type",
			Issue:   "required",
			Message: "El tipo de contenido MIME es obligatorio",
		}
	}

	if !isValidImageContentType(contentType) {
		return nil, &ValidationError{
			Field:   "content_type",
			Issue:   "invalid_format",
			Message: "El formato de imagen no está permitido. Formatos soportados: JPEG, PNG",
		}
	}

	fileSize := req.GetFileSize()
	if fileSize <= 0 {
		return nil, &ValidationError{
			Field:   "file_size",
			Issue:   "invalid_size",
			Message: "El tamaño del archivo debe ser mayor a 0 bytes",
		}
	}

	if fileSize > models.MaxImageSizeBytes {
		return nil, &ValidationError{
			Field:   "file_size",
			Issue:   "size_exceeded",
			Message: "El archivo supera el tamaño máximo permitido de 5 MB",
		}
	}

	// Generar identificador único y clave de objeto según estándar de docs/infrastructure.md
	mediaID := uuid.New().String()
	ext := getExtensionForContentType(contentType, req.GetFileName())

	var objectKey string
	switch resourceType {
	case models.ResourceTypeProfilePhoto:
		objectKey = fmt.Sprintf("profiles/%s/%s%s", ownerID, mediaID, ext)
	case models.ResourceTypeBanner:
		objectKey = fmt.Sprintf("banners/%s/%s%s", ownerID, mediaID, ext)
	case models.ResourceTypePostAttachment:
		objectKey = fmt.Sprintf("posts/%s/%s%s", ownerID, mediaID, ext)
	}

	// Generar URL prefirmada de subida (validez 15 minutos)
	expiresIn := 900 // 15 minutos
	uploadURL, err := s.storageRepo.GeneratePresignedUploadURL(ctx, objectKey, contentType, time.Duration(expiresIn)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("error al generar URL de carga: %w", err)
	}

	return &models.PresignedUploadResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		MediaID:   mediaID,
		ExpiresIn: expiresIn,
	}, nil
}

// VerifyMedia verifica la existencia física de un objeto en el almacenamiento (CU-03)
func (s *mediaService) VerifyMedia(ctx context.Context, objectKey string) (*models.VerifyMediaResponse, error) {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return nil, &ValidationError{
			Field:   "object_key",
			Issue:   "required",
			Message: "La clave del objeto es obligatoria para la verificación",
		}
	}

	exists, size, contentType, err := s.storageRepo.ObjectExists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStorageUnreachable, err)
	}

	if !exists {
		return &models.VerifyMediaResponse{
			Exists:    false,
			ObjectKey: key,
		}, ErrObjectNotFound
	}

	return &models.VerifyMediaResponse{
		Exists:      true,
		ObjectKey:   key,
		ContentType: contentType,
		SizeBytes:   size,
	}, nil
}

// CreatePresignedDownload genera una URL GET temporal para un objeto multimedia existente.
func (s *mediaService) CreatePresignedDownload(ctx context.Context, userID, objectKey string) (*models.PresignedDownloadResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}

	key := strings.TrimSpace(objectKey)
	if key == "" {
		return nil, &ValidationError{
			Field:   "object_key",
			Issue:   "required",
			Message: "La clave del objeto es obligatoria para generar la URL de descarga",
		}
	}
	if len(key) > 255 || strings.Contains(key, "\\") || path.Clean(key) != key ||
		!(strings.HasPrefix(key, "profiles/") ||
			strings.HasPrefix(key, "banners/") ||
			strings.HasPrefix(key, "posts/")) {
		return nil, &ValidationError{
			Field:   "object_key",
			Issue:   "invalid_path",
			Message: "La clave no pertenece a una ruta multimedia permitida",
		}
	}

	exists, _, _, err := s.storageRepo.ObjectExists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStorageUnreachable, err)
	}
	if !exists {
		return nil, ErrObjectNotFound
	}

	const expiresIn = 900
	downloadURL, err := s.storageRepo.GeneratePresignedDownloadURL(ctx, key, time.Duration(expiresIn)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("error al generar URL de descarga: %w", err)
	}

	return &models.PresignedDownloadResponse{
		DownloadURL: downloadURL,
		ObjectKey:   key,
		ExpiresIn:   expiresIn,
	}, nil
}

func isValidImageContentType(ct string) bool {
	switch ct {
	case "image/jpeg", "image/jpg", "image/png":
		return true
	default:
		return false
	}
}

func getExtensionForContentType(ct, filename string) string {
	if filename != "" {
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			return ext
		}
	}

	switch ct {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	default:
		return ""
	}
}
