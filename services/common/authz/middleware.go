package authz

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

const (
	// ContextUserIDKey es la clave de contexto que contiene la identidad validada por autenticación.
	ContextUserIDKey = "user_id"
	// ContextUserRoleKey es la clave de contexto que contiene el rol vigente del usuario autenticado.
	ContextUserRoleKey = "user_role"
)

// RequireRoles devuelve un middleware que solo continúa cuando la petición
// trae una identidad autenticada en el contexto y un rol permitido.
// No consulta cabeceras HTTP: cualquier valor de rol debe haber sido
// previamente validado y almacenado en el contexto por un middleware de
// autenticación.
func RequireRoles(allowed ...string) gin.HandlerFunc {
	allowedRoles := make(map[string]struct{}, len(allowed))
	for _, role := range allowed {
		if canonical, ok := NormalizeRole(role); ok {
			allowedRoles[canonical] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		userIDValue, exists := c.Get(ContextUserIDKey)
		userID, isString := userIDValue.(string)
		if !exists || !isString || strings.TrimSpace(userID) == "" {
			httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
			c.Abort()
			return
		}

		roleValue, exists := c.Get(ContextUserRoleKey)
		rawRole, isString := roleValue.(string)
		if !exists || !isString {
			httpresponse.ForbiddenGin(c, "Rol no válido o no autorizado")
			c.Abort()
			return
		}

		role, valid := NormalizeRole(rawRole)
		if !valid {
			httpresponse.ForbiddenGin(c, "Rol no válido o no autorizado")
			c.Abort()
			return
		}

		if _, permitted := allowedRoles[role]; !permitted {
			httpresponse.ForbiddenGin(c, "No tienes permisos suficientes para esta operación")
			c.Abort()
			return
		}

		c.Next()
	}
}
