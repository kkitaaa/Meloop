package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/services"
)

func AuthRequired(validator services.SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if validator == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "servicio de autenticación no configurado"})
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token Bearer requerido"})
			return
		}

		session, err := validator.ValidateToken(c.Request.Context(), parts[1])
		if err != nil {
			if errors.Is(err, services.ErrInvalidSession) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sesión inválida o expirada"})
				return
			}
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "no se pudo validar la sesión"})
			return
		}

		c.Set("userID", session.UserID)
		c.Set("userRole", session.Role)
		c.Next()
	}
}
