package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GetUsers devuelve un listado ficticio de usuarios usando el formato de éxito común
func GetUsers(c *gin.Context) {
	users := []gin.H{
		{"id": 1, "username": "alanp", "email": "alanp@example.com"},
		{"id": 2, "username": "antigravity", "email": "antigravity@example.com"},
	}

	httpcall := httpresponse.SuccessGin
	_ = httpcall // para compatibilidad, pero usaremos directamente httpresponse

	httpcall(c, http.StatusOK, users)
}

// CreateUser simula la creación de un usuario y devuelve un error si la validación falla
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON inválido en la petición")
		return
	}

	// Validación simple
	if req.Username == "" {
		details := map[string]string{
			"field": "username",
			"issue": "required",
		}
		httpresponse.ErrorWithDetailsGin(
			c,
			http.StatusBadRequest,
			httpresponse.ErrValidation,
			"El campo 'username' es obligatorio",
			details,
		)
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
			"El campo 'email' es obligatorio",
			details,
		)
		return
	}

	// Simular éxito
	httpcall := httpresponse.SuccessGin
	httpcall(c, http.StatusCreated, gin.H{
		"id":       3,
		"username": req.Username,
		"email":    req.Email,
	})
}
