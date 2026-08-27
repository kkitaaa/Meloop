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

// LoginRequest defines the request body for Login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login is the mock login handler originally defined in the codebase
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	if req.Email == "" {
		details := map[string]string{
			"field": "email",
			"issue": "required",
		}
		httpresponse.ErrorWithDetailsGin(
			c,
			http.StatusBadRequest,
			httpcallValidation,
			"El correo electrónico es obligatorio",
			details,
		)
		return
	}

	if req.Password == "" {
		details := map[string]string{
			"field": "password",
			"issue": "required",
		}
		httpresponse.ErrorWithDetailsGin(
			c,
			http.StatusBadRequest,
			httpcallValidation,
			"La contraseña es obligatoria",
			details,
		)
		return
	}

	if req.Password != "meloop123" {
		httpresponse.UnauthorizedGin(c, "Credenciales incorrectas")
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"token":      "ejemplo-token-jwt-seguro",
		"expires_in": 3600,
	})
}
