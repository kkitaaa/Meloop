package controllers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/notification-service/repositories"
	"github.com/meloop/notification-service/services"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type NotificationController struct {
	service *services.InboxService
	logger  *slog.Logger
}

func NewNotificationController(service *services.InboxService) *NotificationController {
	return &NotificationController{
		service: service,
		logger:  logging.New("notification-controller"),
	}
}

func (ctrl *NotificationController) List(c *gin.Context) {
	limit, err := queryInteger(c, "limit", defaultLimit)
	if err != nil || limit < 1 || limit > maxLimit {
		httpresponse.BadRequestGin(c, "El parámetro limit debe estar entre 1 y 100")
		return
	}
	offset, err := queryInteger(c, "offset", 0)
	if err != nil || offset < 0 {
		httpresponse.BadRequestGin(c, "El parámetro offset debe ser un entero no negativo")
		return
	}

	notifications, err := ctrl.service.List(c.Request.Context(), c.GetString(ContextUserIDKey), limit, offset)
	if err != nil {
		ctrl.logger.Error("notification_list_failed", "error", err)
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
	})
}

func (ctrl *NotificationController) MarkRead(c *gin.Context) {
	notificationID := c.Param("id")
	if err := ctrl.service.MarkRead(c.Request.Context(), c.GetString(ContextUserIDKey), notificationID); err != nil {
		if errors.Is(err, repositories.ErrNotificationNotFound) {
			httpresponse.NotFoundGin(c, "No se encontró la notificación")
			return
		}
		ctrl.logger.Error("notification_mark_read_failed", "error", err)
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"id_notificacion": notificationID,
		"leida":           true,
	})
}

func (ctrl *NotificationController) MarkAllRead(c *gin.Context) {
	count, err := ctrl.service.MarkAllRead(c.Request.Context(), c.GetString(ContextUserIDKey))
	if err != nil {
		ctrl.logger.Error("notifications_mark_all_read_failed", "error", err)
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{"actualizadas": count})
}

func queryInteger(c *gin.Context, name string, defaultValue int) (int, error) {
	value := c.Query(name)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}
