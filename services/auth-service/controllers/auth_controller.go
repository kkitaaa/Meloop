package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

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
			httpresponse.ErrValidation,
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
		httpcall := httpresponse.ErrorWithDetailsGin
		httpcall(
			c,
			http.StatusBadRequest,
			httpresponse.ErrValidation,
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
