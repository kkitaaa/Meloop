package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type AdminManager interface {
	ListUsers(context.Context) ([]models.AdminUser, error)
	SetSuspension(context.Context, string, string, bool) (*models.AdminUser, error)
	SetModerator(context.Context, string, string, bool) (*models.AdminUser, error)
}

type AdminController struct {
	service AdminManager
}

func NewAdminController(service AdminManager) *AdminController {
	return &AdminController{service: service}
}

func (ctrl *AdminController) ListUsers(c *gin.Context) {
	users, err := ctrl.service.ListUsers(c.Request.Context())
	if err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, users)
}

func (ctrl *AdminController) SetSuspension(c *gin.Context) {
	var request models.SetSuspensionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.BadRequestGin(c, "Se requiere el campo booleano suspendido")
		return
	}
	user, err := ctrl.service.SetSuspension(c.Request.Context(), c.GetString(ContextUserIDKey), c.Param("id"), request.Suspended)
	ctrl.writeAdminResult(c, user, err)
}

func (ctrl *AdminController) SetModerator(c *gin.Context) {
	var request models.SetModeratorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.BadRequestGin(c, "Se requiere el campo booleano habilitado")
		return
	}
	user, err := ctrl.service.SetModerator(c.Request.Context(), c.GetString(ContextUserIDKey), c.Param("id"), request.Enabled)
	ctrl.writeAdminResult(c, user, err)
}

func (ctrl *AdminController) writeAdminResult(c *gin.Context, user *models.AdminUser, err error) {
	if errors.Is(err, services.ErrAdminUserNotFound) {
		httpresponse.NotFoundGin(c, "Usuario no encontrado")
		return
	}
	if errors.Is(err, services.ErrCannotManageSelf) {
		httpresponse.ErrorGin(c, http.StatusBadRequest, "SELF_ADMIN_CHANGE", "No puedes modificar los controles de tu propia cuenta")
		return
	}
	if err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, user)
}
