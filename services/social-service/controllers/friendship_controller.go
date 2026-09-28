package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/services"
)

type FriendshipController struct {
	service services.FriendshipService
}

func NewFriendshipController(service services.FriendshipService) *FriendshipController {
	return &FriendshipController{service: service}
}

func (ctrl *FriendshipController) Send(c *gin.Context) {
	var request models.SendFriendRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.BadRequestGin(c, "El campo id_usuario es obligatorio")
		return
	}
	result, err := ctrl.service.SendRequest(c, c.GetString(ContextUserIDKey), request.ReceiverID)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusCreated, result)
}

func (ctrl *FriendshipController) Received(c *gin.Context) {
	result, err := ctrl.service.ReceivedRequests(c, c.GetString(ContextUserIDKey))
	if err != nil {
		ctrl.handleError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func (ctrl *FriendshipController) Sent(c *gin.Context) {
	result, err := ctrl.service.SentRequests(c, c.GetString(ContextUserIDKey))
	if err != nil {
		ctrl.handleError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func (ctrl *FriendshipController) Accept(c *gin.Context) {
	ctrl.process(c, func(id int, userID string) (*models.FriendRequest, error) {
		return ctrl.service.AcceptRequest(c, id, userID)
	})
}

func (ctrl *FriendshipController) Reject(c *gin.Context) {
	ctrl.process(c, func(id int, userID string) (*models.FriendRequest, error) {
		return ctrl.service.RejectRequest(c, id, userID)
	})
}

func (ctrl *FriendshipController) Cancel(c *gin.Context) {
	ctrl.process(c, func(id int, userID string) (*models.FriendRequest, error) {
		return ctrl.service.CancelRequest(c, id, userID)
	})
}

func (ctrl *FriendshipController) ListFriends(c *gin.Context) {
	result, err := ctrl.service.ListFriends(c, c.GetString(ContextUserIDKey))
	if err != nil {
		ctrl.handleError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func (ctrl *FriendshipController) RemoveFriend(c *gin.Context) {
	targetID := strings.TrimSpace(c.Param("id"))
	if targetID == "" {
		httpresponse.BadRequestGin(c, "El identificador del amigo o amistad es requerido")
		return
	}
	err := ctrl.service.RemoveFriend(c, c.GetString(ContextUserIDKey), targetID)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{"mensaje": "Amistad eliminada correctamente"})
}

func (ctrl *FriendshipController) GetFriendProfile(c *gin.Context) {
	targetID := strings.TrimSpace(c.Param("id"))
	if targetID == "" {
		httpresponse.BadRequestGin(c, "El identificador del usuario es requerido")
		return
	}
	result, err := ctrl.service.GetFriendProfile(c, c.GetString(ContextUserIDKey), targetID)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func (ctrl *FriendshipController) BlockUser(c *gin.Context) {
	var request models.BlockUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.BadRequestGin(c, "El campo id_usuario es obligatorio")
		return
	}
	err := ctrl.service.BlockUser(c, c.GetString(ContextUserIDKey), request.BlockedID)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{"mensaje": "Usuario bloqueado correctamente"})
}

func (ctrl *FriendshipController) ValidateInteraction(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	targetID := strings.TrimSpace(c.Query("id_usuario"))
	if targetID == "" {
		targetID = strings.TrimSpace(c.Query("target_user_id"))
	}
	if targetID == "" {
		var req models.BlockUserRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			targetID = strings.TrimSpace(req.BlockedID)
		}
	}
	if targetID == "" {
		httpresponse.BadRequestGin(c, "El identificador del usuario objetivo es requerido")
		return
	}

	err := ctrl.service.ValidateInteraction(c, userID, targetID)
	if err != nil {
		if errors.Is(err, services.ErrBlocked) {
			httpresponse.ForbiddenGin(c, "No se permiten interacciones entre estos usuarios debido a un bloqueo")
			return
		}
		ctrl.handleError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, models.InteractionValidationResponse{
		Allowed: true,
	})
}

func (ctrl *FriendshipController) process(c *gin.Context, action func(int, string) (*models.FriendRequest, error)) {
	requestID, err := strconv.Atoi(c.Param("id"))
	if err != nil || requestID <= 0 {
		httpresponse.BadRequestGin(c, "El identificador de la solicitud no es válido")
		return
	}
	result, err := action(requestID, c.GetString(ContextUserIDKey))
	if err != nil {
		ctrl.handleError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func (ctrl *FriendshipController) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrSelfRequest):
		httpresponse.BadRequestGin(c, "No puedes enviarte una solicitud de amistad a ti mismo")
	case errors.Is(err, services.ErrSelfBlock):
		httpresponse.BadRequestGin(c, "No puedes bloquearte a ti mismo")
	case errors.Is(err, services.ErrInvalidUser):
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
	case errors.Is(err, services.ErrRequestConflict):
		httpresponse.ConflictGin(c, "Ya existe una solicitud o amistad entre estos usuarios")
	case errors.Is(err, services.ErrAlreadyBlocked):
		httpresponse.ConflictGin(c, "El usuario ya se encuentra bloqueado")
	case errors.Is(err, services.ErrBlocked):
		httpresponse.ForbiddenGin(c, "No se permiten interacciones entre estos usuarios debido a un bloqueo")
	case errors.Is(err, services.ErrRequestNotFound):
		httpresponse.NotFoundGin(c, "Solicitud de amistad no encontrada")
	case errors.Is(err, services.ErrFriendshipNotFound):
		httpresponse.NotFoundGin(c, "Relación de amistad no encontrada")
	case errors.Is(err, services.ErrUserNotFound):
		httpresponse.NotFoundGin(c, "Usuario no encontrado")
	case errors.Is(err, services.ErrNotFriends):
		httpresponse.ForbiddenGin(c, "No existe una relación de amistad con este usuario")
	case errors.Is(err, services.ErrNotReceiver), errors.Is(err, services.ErrNotSender), errors.Is(err, services.ErrNotAuthorized):
		httpresponse.ForbiddenGin(c, "No tienes autorización para realizar esta acción")
	case errors.Is(err, services.ErrRequestProcessed):
		httpresponse.ConflictGin(c, "La solicitud de amistad ya fue procesada")
	default:
		httpresponse.InternalErrorGin(c)
	}
}
