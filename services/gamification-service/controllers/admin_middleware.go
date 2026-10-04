package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/sessionauth"
)

const adminUserIDKey = "admin_user_id"

type AdminRoleRepository interface {
	IsAdmin(context.Context, string) (bool, error)
}

func AdminRequired(authServiceURL string, repository AdminRoleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := sessionauth.UserID(c.Request.Context(), authServiceURL, c.GetHeader("Authorization"))
		if errors.Is(err, sessionauth.ErrUnauthorized) {
			httpresponse.UnauthorizedGin(c, "Se requiere una sesión válida")
			c.Abort()
			return
		}
		if errors.Is(err, sessionauth.ErrUnavailable) {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "AUTH_SERVICE_UNAVAILABLE", "No se pudo validar la sesión")
			c.Abort()
			return
		}
		if err != nil {
			httpresponse.InternalErrorGin(c)
			c.Abort()
			return
		}
		isAdmin, err := repository.IsAdmin(c.Request.Context(), userID)
		if err != nil {
			httpresponse.InternalErrorGin(c)
			c.Abort()
			return
		}
		if !isAdmin {
			httpresponse.ErrorGin(c, http.StatusForbidden, "ADMIN_REQUIRED", "Se requiere el rol de administrador")
			c.Abort()
			return
		}
		c.Set(adminUserIDKey, userID)
		c.Next()
	}
}
