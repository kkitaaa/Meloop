package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/services"
)

const (
	// ContextUserIDKey es la clave utilizada en el contexto de Gin para almacenar el ID del usuario
	ContextUserIDKey = "user_id"
	// ContextUserRoleKey es la clave utilizada para conservar el rol del usuario autenticado
	ContextUserRoleKey = "user_role"
)

// AuthRequired es un middleware que verifica la presencia del identificador de usuario en el contexto de autenticación
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Verificar si ya fue fijado en el contexto por un middleware previo o prueba
		if val, exists := c.Get(ContextUserIDKey); exists {
			if userID, ok := val.(string); ok && strings.TrimSpace(userID) != "" {
				c.Set(ContextUserIDKey, strings.TrimSpace(userID))
				c.Next()
				return
			}
		}

		// 2. Verificar si viene en los encabezados estándar de propagación de identidad
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = c.GetHeader("X-User-Id")
		}

		userID = strings.TrimSpace(userID)
		if userID == "" {
			httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, userID)
		if role := strings.TrimSpace(c.GetHeader("X-User-Role")); role != "" {
			c.Set(ContextUserRoleKey, strings.ToLower(role))
		}
		c.Next()
	}
}

func AuthRequiredWithSessionValidation(validator services.SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := validator.ValidateSession(c.Request.Context(), c.GetHeader("Authorization"))
		if err != nil {
			if errors.Is(err, services.ErrInvalidSession) {
				httpresponse.UnauthorizedGin(c, "Sesión inválida o expirada")
				c.Abort()
				return
			}
			httpresponse.ErrorGin(
				c,
				http.StatusServiceUnavailable,
				"AUTH_SERVICE_UNAVAILABLE",
				"No se pudo validar la sesión de usuario",
			)
			c.Abort()
			return
		}
		if strings.TrimSpace(userID) == "" {
			httpresponse.UnauthorizedGin(c, "Sesión inválida o expirada")
			c.Abort()
			return
		}
		c.Set(ContextUserIDKey, strings.TrimSpace(userID))
		c.Next()
	}
}

// AdminRequired valida que la sesión actual tenga permisos de administrador o moderación privilegiada.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue := strings.TrimSpace(c.GetHeader("X-User-Role"))
		if roleValue == "" {
			roleValue = strings.TrimSpace(c.GetHeader("X-Role"))
		}
		if roleValue == "" {
			if value, exists := c.Get(ContextUserRoleKey); exists {
				if role, ok := value.(string); ok {
					roleValue = role
				}
			}
		}
		if roleValue == "" {
			httpresponse.ForbiddenGin(c, "Se requieren permisos de administrador")
			c.Abort()
			return
		}

		roles := strings.FieldsFunc(strings.ToLower(roleValue), func(r rune) bool {
			return r == ',' || r == ';' || r == '|' || r == ' '
		})
		isAllowed := false
		for _, role := range roles {
			if role == "admin" || role == "administrator" || role == "administrador" || role == "moderator" || role == "moderador" {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			httpresponse.ForbiddenGin(c, "Se requieren permisos de administrador")
			c.Abort()
			return
		}

		c.Set(ContextUserRoleKey, strings.ToLower(roleValue))
		c.Next()
	}
}
