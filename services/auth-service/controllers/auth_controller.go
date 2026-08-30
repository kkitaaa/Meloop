package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/services"
	"github.com/meloop/services/common/httpresponse"
)

// AuthController handles authentication HTTP requests
type AuthController struct {
	authService services.AuthService
}

// NewAuthController creates a new AuthController instance
func NewAuthController(srv services.AuthService) *AuthController {
	return &AuthController{authService: srv}
}

// Register handles POST /auth/register
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

// Login handles POST /auth/login
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
