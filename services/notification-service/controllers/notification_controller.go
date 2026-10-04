package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/notification-service/models"
	"github.com/meloop/services/common/httpresponse"
)

type NotificationStore interface {
	List(context.Context, string, int, int) ([]models.NotificationRecord, error)
	MarkRead(context.Context, string, string) (bool, error)
	MarkAllRead(context.Context, string) (int64, error)
	ListPreferences(context.Context, string) ([]models.NotificationPreference, error)
	SetPreference(context.Context, string, string, bool) error
}

type NotificationController struct {
	store NotificationStore
}

func NewNotificationController(store NotificationStore) *NotificationController {
	return &NotificationController{store: store}
}

func (ctrl *NotificationController) List(c *gin.Context) {
	limit, err := queryInt(c, "limit", 50)
	if err != nil || limit < 1 || limit > 100 {
		httpresponse.ErrorGin(c, http.StatusBadRequest, "INVALID_LIMIT", "limit debe ser un número entre 1 y 100")
		return
	}
	offset, err := queryInt(c, "offset", 0)
	if err != nil || offset < 0 {
		httpresponse.ErrorGin(c, http.StatusBadRequest, "INVALID_OFFSET", "offset debe ser un entero no negativo")
		return
	}

	notifications, err := ctrl.store.List(c.Request.Context(), c.GetString("user_id"), limit, offset)
	if err != nil {
		httpresponse.ErrorGin(c, http.StatusInternalServerError, "NOTIFICATIONS_LIST_FAILED", "No se pudieron consultar las notificaciones")
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"items":  notifications,
		"limit":  limit,
		"offset": offset,
	})
}

func (ctrl *NotificationController) MarkRead(c *gin.Context) {
	updated, err := ctrl.store.MarkRead(c.Request.Context(), c.GetString("user_id"), c.Param("id"))
	if err != nil {
		httpresponse.ErrorGin(c, http.StatusInternalServerError, "NOTIFICATION_UPDATE_FAILED", "No se pudo actualizar la notificación")
		return
	}
	if !updated {
		httpresponse.ErrorGin(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "No se encontró la notificación")
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{"read": true})
}

func (ctrl *NotificationController) MarkAllRead(c *gin.Context) {
	count, err := ctrl.store.MarkAllRead(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		httpresponse.ErrorGin(c, http.StatusInternalServerError, "NOTIFICATIONS_UPDATE_FAILED", "No se pudieron actualizar las notificaciones")
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{"updated": count})
}

func (ctrl *NotificationController) GetPreferences(c *gin.Context) {
	preferences, err := ctrl.store.ListPreferences(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		httpresponse.ErrorGin(c, http.StatusInternalServerError, "NOTIFICATION_PREFERENCES_FAILED", "No se pudieron consultar las preferencias")
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, preferences)
}

func (ctrl *NotificationController) SetPreference(c *gin.Context) {
	notificationType := c.Param("type")
	if !models.IsSupportedNotificationType(notificationType) {
		httpresponse.ErrorGin(c, http.StatusBadRequest, "INVALID_NOTIFICATION_TYPE", "El tipo de notificación no es válido")
		return
	}
	var request struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Enabled == nil {
		httpresponse.ErrorGin(c, http.StatusBadRequest, "INVALID_PREFERENCE", "Se requiere el campo booleano enabled")
		return
	}
	if err := ctrl.store.SetPreference(c.Request.Context(), c.GetString("user_id"), notificationType, *request.Enabled); err != nil {
		httpresponse.ErrorGin(c, http.StatusInternalServerError, "NOTIFICATION_PREFERENCE_UPDATE_FAILED", "No se pudo actualizar la preferencia")
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, models.NotificationPreference{Type: notificationType, Enabled: *request.Enabled})
}

func queryInt(c *gin.Context, key string, defaultValue int) (int, error) {
	value := c.Query(key)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}
