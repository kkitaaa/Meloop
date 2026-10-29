package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

const contextUserRoleKey = "user_role"

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue := strings.TrimSpace(c.GetHeader("X-User-Role"))
		if roleValue == "" {
			roleValue = strings.TrimSpace(c.GetHeader("X-Role"))
		}
		if roleValue == "" {
			if value, exists := c.Get(contextUserRoleKey); exists {
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

		c.Set(contextUserRoleKey, strings.ToLower(roleValue))
		c.Next()
	}
}
