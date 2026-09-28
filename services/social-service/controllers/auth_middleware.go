package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

const ContextUserIDKey = "user_id"

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if value, exists := c.Get(ContextUserIDKey); exists {
			if userID, ok := value.(string); ok && strings.TrimSpace(userID) != "" {
				c.Set(ContextUserIDKey, strings.TrimSpace(userID))
				c.Next()
				return
			}
		}

		userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
		if userID == "" {
			httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
			c.Abort()
			return
		}
		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}
