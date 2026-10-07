package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

// ProfileController gestiona el endpoint HTTP de edición de perfil (RF-07)
type ProfileController struct {
	profileService services.ProfileService
}

// NewProfileController crea una nueva instancia de ProfileController
func NewProfileController(profileService services.ProfileService) *ProfileController {
	return &ProfileController{profileService: profileService}
}

// UpdateProfile maneja la actualización del perfil del usuario autenticado (PUT /v1/users/me/profile - RF-07)
func (ctrl *ProfileController) UpdateProfile(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	updated, err := ctrl.profileService.UpdateProfile(c.Request.Context(), userID, &req)
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

		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}

		if errors.Is(err, services.ErrMediaStorageFailed) {
			httpresponse.ErrorGin(
				c,
				http.StatusBadGateway,
				httpresponse.ErrInternal,
				"Error en el almacenamiento de objetos multimedia. La actualización del perfil fue cancelada para preservar la consistencia.",
			)
			return
		}

		if errors.Is(err, services.ErrMediaNotFound) {
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpresponse.ErrValidation,
				"El archivo multimedia especificado no existe en el almacenamiento",
				map[string]string{"issue": "media_not_found"},
			)
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, updated)
}
