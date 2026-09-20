package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

// NotificationController gestiona los endpoints HTTP de configuración de notificaciones (RF-05 / RF-52)
type NotificationController struct {
	notifService services.NotificationConfigService
}

// NewNotificationController crea una nueva instancia de NotificationController
func NewNotificationController(notifService services.NotificationConfigService) *NotificationController {
	return &NotificationController{notifService: notifService}
}

// GetNotificationSettings maneja la consulta de configuraciones de notificación (GET /users/me/notifications/settings)
func (ctrl *NotificationController) GetNotificationSettings(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	res, err := ctrl.notifService.GetSettings(c.Request.Context(), userID)
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

// UpdateNotificationSettings maneja la actualización individual o masiva de configuraciones (PATCH /users/me/notifications/settings)
func (ctrl *NotificationController) UpdateNotificationSettings(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	bodyBytes, err := c.GetRawData()
	if err != nil || len(bodyBytes) == 0 {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	// 1. Intentar deserializar como un array directo de items
	var listReq []models.NotificationSettingItem
	if err := json.Unmarshal(bodyBytes, &listReq); err == nil && len(listReq) > 0 {
		ctrl.handleBatchUpdate(c, userID, listReq)
		return
	}

	// 2. Intentar deserializar como objeto contenedor {"configuraciones": [...] / "settings": [...]}
	var batchReq models.UpdateNotificationsBatchRequest
	if err := json.Unmarshal(bodyBytes, &batchReq); err == nil && len(batchReq.GetItems()) > 0 {
		ctrl.handleBatchUpdate(c, userID, batchReq.GetItems())
		return
	}

	// 3. Intentar deserializar como objeto de item individual {"tipo_notificacion": "LIKE", "habilitada": false}
	var singleReq models.UpdateSingleNotificationRequest
	if err := json.Unmarshal(bodyBytes, &singleReq); err == nil && singleReq.TipoNotificacion != "" {
		if singleReq.Habilitada == nil {
			details := map[string]string{
				"field": "habilitada",
				"issue": "required",
			}
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpresponse.ErrValidation,
				"El campo 'habilitada' es obligatorio",
				details,
			)
			return
		}

		res, err := ctrl.notifService.UpdateSetting(c.Request.Context(), userID, singleReq.TipoNotificacion, *singleReq.Habilitada)
		if err != nil {
			ctrl.handleServiceError(c, err)
			return
		}

		httpresponse.SuccessGin(c, http.StatusOK, res)
		return
	}

	httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
}

// UpdateSingleSetting maneja la actualización indicando el tipo en el path (PATCH /users/me/notifications/settings/:tipo)
func (ctrl *NotificationController) UpdateSingleSetting(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	tipo := c.Param("tipo")
	if tipo == "" {
		httpresponse.BadRequestGin(c, "Tipo de notificación no especificado")
		return
	}

	var req models.UpdateSingleNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	if req.Habilitada == nil {
		details := map[string]string{
			"field": "habilitada",
			"issue": "required",
		}
		httpresponse.ErrorWithDetailsGin(
			c,
			http.StatusBadRequest,
			httpresponse.ErrValidation,
			"El campo 'habilitada' es obligatorio",
			details,
		)
		return
	}

	res, err := ctrl.notifService.UpdateSetting(c.Request.Context(), userID, tipo, *req.Habilitada)
	if err != nil {
		ctrl.handleServiceError(c, err)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, res)
}

func (ctrl *NotificationController) handleBatchUpdate(c *gin.Context, userID string, items []models.NotificationSettingItem) {
	res, err := ctrl.notifService.UpdateBatch(c.Request.Context(), userID, items)
	if err != nil {
		ctrl.handleServiceError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, res)
}

func (ctrl *NotificationController) handleServiceError(c *gin.Context, err error) {
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
}
