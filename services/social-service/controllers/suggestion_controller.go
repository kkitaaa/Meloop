package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/services"
)

type SuggestionController struct {
	service services.FriendshipService
}

func NewSuggestionController(service services.FriendshipService) *SuggestionController {
	return &SuggestionController{service: service}
}

func (ctrl *SuggestionController) GetFriendSuggestions(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString(ContextUserIDKey))
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	result, err := ctrl.service.GetFriendSuggestions(c.Request.Context(), userID)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	if result == nil {
		result = []models.FriendSuggestion{}
	}

	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func (ctrl *SuggestionController) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidUser):
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
	case errors.Is(err, services.ErrNoMusicalPreferences):
		httpresponse.BadRequestGin(c, "La cuenta no posee preferencias musicales registradas para realizar el cruce")
	case errors.Is(err, services.ErrInsufficientCandidates), errors.Is(err, services.ErrNoCandidates):
		httpresponse.NotFoundGin(c, "No se encontraron candidatos disponibles para sugerencias de amistad")
	case errors.Is(err, services.ErrBlocked):
		httpresponse.ForbiddenGin(c, "No se permiten interacciones entre estos usuarios debido a un bloqueo")
	case errors.Is(err, services.ErrUserNotFound):
		httpresponse.NotFoundGin(c, "Usuario no encontrado")
	default:
		httpresponse.InternalErrorGin(c)
	}
}
