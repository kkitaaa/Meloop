package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

// AuthRequired is a middleware that requires a valid session token
func (ctrl *AuthController) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httpresponse.UnauthorizedGin(c, "Token de sesión no proporcionado")
			c.Abort()
			return
		}

		// Support both "Bearer <token>" and just "<token>"
		token := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			if len(authHeader) > 7 {
				token = authHeader[7:]
			} else {
				token = ""
			}
		}
		token = strings.TrimSpace(token)

		if token == "" {
			httpresponse.UnauthorizedGin(c, "Token de sesión vacío")
			c.Abort()
			return
		}

		user, err := ctrl.authService.ValidateSession(c.Request.Context(), token)
		if err != nil {
			httpcall := httpresponse.UnauthorizedGin
			httpcall(c, "Sesión inválida o expirada")
			c.Abort()
			return
		}

		// Store user info in context for downstream handlers
		c.Set("user", user)
		c.Set("session_token", token)
		c.Next()
	}
}
