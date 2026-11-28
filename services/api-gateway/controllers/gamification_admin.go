package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/authz"
)

const gamificationAdminPath = "/api/v1/gamification/admin"

// RequireGamificationAdmin aplica autenticación y autorización de administrador
// únicamente a los endpoints administrativos de gamificación. El resto de rutas
// de gamificación continúa con el enrutamiento existente sin cambios.
func RequireGamificationAdmin() gin.HandlerFunc {
	requireAdmin := authz.RequireRoles(authz.RoleAdmin)

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path != gamificationAdminPath && !strings.HasPrefix(path, gamificationAdminPath+"/") {
			c.Next()
			return
		}

		if !authenticateSession(c) {
			return
		}
		requireAdmin(c)
	}
}
