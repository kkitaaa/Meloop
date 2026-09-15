package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

const (
	// ContextUserIDKey es la clave utilizada en el contexto de Gin para almacenar el ID del usuario
	ContextUserIDKey = "user_id"
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
		c.Next()
	}
}
