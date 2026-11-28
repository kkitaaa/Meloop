package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type MusicPreferencesController struct {
	service services.MusicPreferencesService
}

func NewMusicPreferencesController(service services.MusicPreferencesService) *MusicPreferencesController {
	return &MusicPreferencesController{service: service}
}

func (ctrl *MusicPreferencesController) Get(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString(ContextUserIDKey))
	result, err := ctrl.service.Get(c.Request.Context(), userID)
	if err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func (ctrl *MusicPreferencesController) Add(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString(ContextUserIDKey))
	var request models.AddMusicPreferenceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de preferencia musical inválido")
		return
	}

	result, err := ctrl.service.Add(c.Request.Context(), userID, &request)
	if err != nil {
		writeMusicPreferenceError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func (ctrl *MusicPreferencesController) Remove(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString(ContextUserIDKey))
	result, err := ctrl.service.Remove(
		c.Request.Context(),
		userID,
		c.Param("type"),
		c.Param("item_id"),
	)
	if err != nil {
		writeMusicPreferenceError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func writeMusicPreferenceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidMusicPreference):
		httpresponse.BadRequestGin(c, "Tipo, ID, nombre o imagen de preferencia inválidos")
	case errors.Is(err, services.ErrPreferenceAlreadyExists):
		httpresponse.ErrorGin(
			c,
			http.StatusConflict,
			"ALREADY_EXISTS",
			"La preferencia musical ya está guardada",
		)
	case errors.Is(err, services.ErrPreferenceLimitReached):
		httpresponse.ErrorGin(
			c,
			http.StatusConflict,
			"PREFERENCE_LIMIT",
			"Se alcanzó el límite de preferencias para este tipo",
		)
	default:
		httpresponse.InternalErrorGin(c)
	}
}
