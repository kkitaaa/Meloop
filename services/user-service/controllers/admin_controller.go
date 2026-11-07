package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

// AdminController expone las operaciones exclusivas para administradores.
type AdminController struct {
	userService services.UserService
}

func NewAdminController(userService services.UserService) *AdminController {
	return &AdminController{userService: userService}
}

func (ctrl *AdminController) ListUsers(c *gin.Context) {
	users, err := ctrl.userService.ListUsers(c.Request.Context())
	if err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, users)
}

func (ctrl *AdminController) SuspendUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		httpresponse.BadRequestGin(c, "Debe indicar el identificador del usuario")
		return
	}

	var req models.UserModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = models.UserModerationRequest{}
	}

	user, err := ctrl.userService.SuspendUser(c.Request.Context(), userID, req.Reason)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}
		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, user)
}

func (ctrl *AdminController) ReactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		httpresponse.BadRequestGin(c, "Debe indicar el identificador del usuario")
		return
	}

	user, err := ctrl.userService.ReactivateUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}
		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, user)
}

func (ctrl *AdminController) UpdateModeratorRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		httpresponse.BadRequestGin(c, "Debe indicar el identificador del usuario")
		return
	}

	var req models.ModeratorPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	user, err := ctrl.userService.SetModeratorRole(c.Request.Context(), userID, req.Moderador)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}
		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, user)
}
