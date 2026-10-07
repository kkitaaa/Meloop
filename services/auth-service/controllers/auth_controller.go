package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/services"
	"github.com/meloop/services/common/httpresponse"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(srv services.AuthService) *AuthController {
	return &AuthController{authService: srv}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	res, err := ctrl.authService.Register(c.Request.Context(), &req)
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
				httpcallFieldIssue(valErr),
				valErr.Message,
				details,
			)
			return
		}

		if errors.Is(err, services.ErrUsernameExists) {
			httpresponse.ConflictGin(c, "El nombre de usuario ya está registrado")
			return
		}

		if errors.Is(err, services.ErrEmailExists) {
			httpcall := httpresponse.ConflictGin
			httpcall(c, "El correo electrónico ya está registrado")
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusCreated, res)
}

func httpcallFieldIssue(err *services.ValidationError) string {
	return httpcallValidation
}

const httpcallValidation = httpresponse.ErrValidation

func (ctrl *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	res, err := ctrl.authService.Login(c.Request.Context(), &req)
	if err != nil {
		var valErr *services.ValidationError
		if errors.As(err, &valErr) {
			details := map[string]string{
				"field": valErr.Field,
				"issue": valErr.Issue,
			}
			httpcall := httpcallValidation
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpcall,
				valErr.Message,
				details,
			)
			return
		}

		if errors.Is(err, services.ErrInvalidCredentials) {
			httpresponse.UnauthorizedGin(c, "Credenciales incorrectas")
			return
		}

		httpcall := httpresponse.InternalErrorGin
		httpcall(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, res)
}

// Logout handles POST /auth/logout
func (ctrl *AuthController) Logout(c *gin.Context) {
	token, exists := c.Get("session_token")
	if !exists {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión activa")
		return
	}

	err := ctrl.authService.Logout(c.Request.Context(), token.(string))
	if err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"message": "Sesión cerrada correctamente",
	})
}

// Validate handles GET /auth/validate
func (ctrl *AuthController) Validate(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión activa")
		return
	}

	httpcall := httpresponse.SuccessGin
	httpcall(c, http.StatusOK, user)
}

// ChangePassword maneja la petición de cambio de contraseña (POST /auth/change-password)
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión activa")
		return
	}

	sessionUser, ok := userVal.(*models.SessionUser)
	if !ok || sessionUser == nil || sessionUser.ID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión activa")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	err := ctrl.authService.ChangePassword(c.Request.Context(), sessionUser.ID, &req)
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
				httpcallValidation,
				valErr.Message,
				details,
			)
			return
		}

		if errors.Is(err, services.ErrInvalidCredentials) {
			httpresponse.UnauthorizedGin(c, "La contraseña actual es incorrecta")
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"message": "Contraseña actualizada correctamente",
	})
}

// RequestPasswordRecovery maneja la petición de inicio de recuperación de contraseña (POST /auth/password-recovery)
func (ctrl *AuthController) RequestPasswordRecovery(c *gin.Context) {
	var req models.PasswordRecoveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	err := ctrl.authService.RequestPasswordRecovery(c.Request.Context(), &req)
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
				httpcallValidation,
				valErr.Message,
				details,
			)
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"message": "Si el correo electrónico está registrado, recibirás un enlace de recuperación",
	})
}

// ResetPassword maneja la petición para restablecer la contraseña mediante token (POST /auth/password-recovery/reset)
func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	err := ctrl.authService.ResetPassword(c.Request.Context(), &req)
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
				httpcallValidation,
				valErr.Message,
				details,
			)
			return
		}

		if errors.Is(err, services.ErrTokenNotFound) {
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpcallValidation,
				"El token de recuperación es inválido o no existe",
				map[string]string{
					"field": "token",
					"issue": "invalid",
				},
			)
			return
		}

		if errors.Is(err, services.ErrTokenExpired) {
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpcallValidation,
				"El token de recuperación ha expirado",
				map[string]string{
					"field": "token",
					"issue": "expired",
				},
			)
			return
		}

		if errors.Is(err, services.ErrTokenAlreadyUsed) {
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpcallValidation,
				"El token de recuperación ya ha sido utilizado",
				map[string]string{
					"field": "token",
					"issue": "already_used",
				},
			)
			return
		}

		if errors.Is(err, services.ErrInvalidCredentials) {
			httpresponse.UnauthorizedGin(c, "Credenciales inválidas")
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"message": "Contraseña restablecida exitosamente",
	})
}
