package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/media-service/models"
	"github.com/meloop/media-service/services"
	"github.com/meloop/services/common/httpresponse"
)

const ContextUserIDKey = "user_id"

// MediaController gestiona los endpoints HTTP de multimedia para foto y banner de perfil (REI-03 / CU-03)
type MediaController struct {
	mediaService services.MediaService
}

// NewMediaController crea una nueva instancia de MediaController
func NewMediaController(mediaService services.MediaService) *MediaController {
	return &MediaController{mediaService: mediaService}
}

// GeneratePresignedUpload genera una URL prefirmada para que el usuario autenticado cargue su foto o banner directamente a MinIO/S3 (REI-03)
func (ctrl *MediaController) GeneratePresignedUpload(c *gin.Context) {
	// 1. Validar autenticación obligatoria del usuario
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
		if userID == "" {
			userID = c.GetHeader("X-User-Id")
		}
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada para solicitar URL prefirmada")
		return
	}

	var req models.PresignedUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	resp, err := ctrl.mediaService.CreatePresignedUpload(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
			return
		}

		var valErr *services.ValidationError
		if errors.As(err, &valErr) {
			details := map[string]string{
				"field": valErr.Field,
				"issue": valErr.Issue,
			}
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpresponse.ErrValidation,
				valErr.Message,
				details,
			)
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, resp)
}

// VerifyMedia verifica si un objeto ya fue subido exitosamente a MinIO/S3 antes de persistirlo en el perfil (CU-03)
func (ctrl *MediaController) VerifyMedia(c *gin.Context) {
	var objectKey string

	if c.Request.Method == http.MethodPost {
		var req models.VerifyMediaRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.ObjectKey != "" {
			objectKey = req.ObjectKey
		}
	}

	if objectKey == "" {
		objectKey = c.Query("key")
		if objectKey == "" {
			objectKey = c.Query("object_key")
		}
	}

	if strings.TrimSpace(objectKey) == "" {
		httpresponse.BadRequestGin(c, "El parámetro 'object_key' o 'key' es obligatorio")
		return
	}

	resp, err := ctrl.mediaService.VerifyMedia(c.Request.Context(), objectKey)
	if err != nil {
		var valErr *services.ValidationError
		if errors.As(err, &valErr) {
			details := map[string]string{
				"field": valErr.Field,
				"issue": valErr.Issue,
			}
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpresponse.ErrValidation,
				valErr.Message,
				details,
			)
			return
		}

		if errors.Is(err, services.ErrObjectNotFound) {
			httpresponse.NotFoundGin(c, "El archivo multimedia no existe en el almacenamiento")
			return
		}

		if errors.Is(err, services.ErrStorageUnreachable) {
			httpresponse.ErrorGin(c, http.StatusBadGateway, httpresponse.ErrInternal, "No se pudo conectar con el almacenamiento MinIO/S3")
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, resp)
}
