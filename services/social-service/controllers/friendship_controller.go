package controllers

import (
	"errors"
	"net/http"
	"strconv"

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
	case errors.Is(err, services.ErrInvalidUser):
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
	case errors.Is(err, services.ErrRequestConflict):
		httpresponse.ConflictGin(c, "Ya existe una solicitud o amistad entre estos usuarios")
	case errors.Is(err, services.ErrBlocked):
		httpresponse.ForbiddenGin(c, "No se puede enviar una solicitud entre estos usuarios")
	case errors.Is(err, services.ErrRequestNotFound):
		httpresponse.NotFoundGin(c, "Solicitud de amistad no encontrada")
	case errors.Is(err, services.ErrNotReceiver), errors.Is(err, services.ErrNotSender):
		httpresponse.ForbiddenGin(c, "No tienes autorización para modificar esta solicitud")
	case errors.Is(err, services.ErrRequestProcessed):
		httpresponse.ConflictGin(c, "La solicitud de amistad ya fue procesada")
	default:
		httpresponse.InternalErrorGin(c)
	}
}
