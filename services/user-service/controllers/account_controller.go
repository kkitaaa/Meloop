package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

// AccountController gestiona las peticiones HTTP relacionadas con la cuenta del usuario (RF-05)
type AccountController struct {
	userService services.UserService
}

// NewAccountController crea una nueva instancia del controlador de cuentas
func NewAccountController(userService services.UserService) *AccountController {
	return &AccountController{userService: userService}
}

// GetAccount maneja la consulta de información básica de la cuenta (GET /users/me)
func (ctrl *AccountController) GetAccount(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	acc, err := ctrl.userService.GetAccount(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}
		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, acc)
}

// UpdateUsername maneja la actualización del nombre de usuario (PATCH /users/me/username)
func (ctrl *AccountController) UpdateUsername(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	var req models.UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	acc, err := ctrl.userService.UpdateUsername(c.Request.Context(), userID, &req)
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

		if errors.Is(err, services.ErrUsernameExists) {
			httpresponse.ConflictGin(c, "El nombre de usuario ya está registrado")
			return
		}

		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, acc)
}

// RequestEmailChange maneja la solicitud de cambio de correo (POST /users/me/email)
func (ctrl *AccountController) RequestEmailChange(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	var req models.RequestEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	res, err := ctrl.userService.RequestEmailChange(c.Request.Context(), userID, &req)
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

		if errors.Is(err, services.ErrEmailSameAsCurrent) {
			details := map[string]string{
				"field": "email",
				"issue": "same_as_current",
			}
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpresponse.ErrValidation,
				"El nuevo correo no puede ser igual al correo actual",
				details,
			)
			return
		}

		if errors.Is(err, services.ErrEmailExists) {
			httpresponse.ConflictGin(c, "El correo electrónico ya está registrado")
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
