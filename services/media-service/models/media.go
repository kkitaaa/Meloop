package models

import "strings"

// Constantes de tipos de recursos permitidos para esta tarea (REI-03)
const (
	ResourceTypeProfilePhoto = "profile_photo"
	ResourceTypeBanner       = "banner"
)

// Límites de tamaño en bytes según el MVP y SRS (REI-03)
const (
	MaxImageSizeBytes = 5 * 1024 * 1024 // 5 MB
)

// PresignedUploadRequest contiene los parámetros para solicitar una URL prefirmada de carga
type PresignedUploadRequest struct {
	ResourceType string `json:"resource_type"`
	TipoRecurso  string `json:"tipo_recurso"` // Alias en español
	ContentType  string `json:"content_type"`
	TipoMIME     string `json:"tipo_mime"` // Alias en español
	MIME         string `json:"mime"`      // Alias
	FileName     string `json:"file_name"`
	Nombre       string `json:"nombre"` // Alias en español
	FileSize     int64  `json:"file_size"`
	TamanoBytes  int64  `json:"tamano_bytes"` // Alias en español
	Size         int64  `json:"size"`         // Alias
}

// GetResourceType normaliza y devuelve el tipo de recurso solicitado
func (r *PresignedUploadRequest) GetResourceType() string {
	res := strings.TrimSpace(r.ResourceType)
	if res == "" {
		res = strings.TrimSpace(r.TipoRecurso)
	}
	res = strings.ToLower(res)
	switch res {
	case "profile_photo", "profile", "profiles", "foto_perfil", "foto", "avatar":
		return ResourceTypeProfilePhoto
	case "banner", "banners", "banner_perfil":
		return ResourceTypeBanner
	default:
		return res
	}
}

// GetContentType normaliza y devuelve el tipo MIME
func (r *PresignedUploadRequest) GetContentType() string {
	ct := strings.TrimSpace(r.ContentType)
	if ct == "" {
		ct = strings.TrimSpace(r.TipoMIME)
	}
	if ct == "" {
		ct = strings.TrimSpace(r.MIME)
	}
	return strings.ToLower(ct)
}

// GetFileName devuelve el nombre de archivo solicitado
func (r *PresignedUploadRequest) GetFileName() string {
	fn := strings.TrimSpace(r.FileName)
	if fn == "" {
		fn = strings.TrimSpace(r.Nombre)
	}
	return fn
}

// GetFileSize devuelve el tamaño en bytes solicitado
func (r *PresignedUploadRequest) GetFileSize() int64 {
	if r.FileSize > 0 {
		return r.FileSize
	}
	if r.TamanoBytes > 0 {
		return r.TamanoBytes
	}
	return r.Size
}

// PresignedUploadResponse representa la respuesta exitosa con la URL prefirmada
type PresignedUploadResponse struct {
	UploadURL string `json:"upload_url"`
	ObjectKey string `json:"object_key"`
	MediaID   string `json:"media_id"`
	ExpiresIn int    `json:"expires_in"` // Segundos de validez
}

// VerifyMediaRequest solicita la verificación de existencia de un objeto en MinIO/S3
type VerifyMediaRequest struct {
	ObjectKey string `json:"object_key"`
}

// VerifyMediaResponse representa el resultado de la verificación
type VerifyMediaResponse struct {
	Exists      bool   `json:"exists"`
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type,omitempty"`
	SizeBytes   int64  `json:"size_bytes,omitempty"`
}
