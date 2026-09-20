package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

// PrivacyController gestiona los endpoints HTTP de configuración de privacidad (RF-05 / RF-55)
type PrivacyController struct {
	privacyService services.PrivacyService
}

// NewPrivacyController crea una nueva instancia de PrivacyController
func NewPrivacyController(privacyService services.PrivacyService) *PrivacyController {
	return &PrivacyController{privacyService: privacyService}
}

// GetPrivacy maneja la consulta de configuración de privacidad del usuario autenticado (GET /users/me/privacy)
func (ctrl *PrivacyController) GetPrivacy(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	res, err := ctrl.privacyService.GetPrivacy(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}
		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, res)
}

// UpdatePrivacy maneja la actualización parcial o total de la configuración de privacidad (PATCH /users/me/privacy)
func (ctrl *PrivacyController) UpdatePrivacy(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	var req models.UpdatePrivacyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	res, err := ctrl.privacyService.UpdatePrivacy(c.Request.Context(), userID, &req)
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

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, res)
}
