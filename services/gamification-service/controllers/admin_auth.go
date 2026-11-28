package controllers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/services"
	"github.com/meloop/services/common/authz"
	"github.com/meloop/services/common/httpresponse"
)

// SessionValidator valida el token de sesión contra auth-service.
type SessionValidator interface {
	ValidateToken(ctx context.Context, token string) (services.ValidatedSession, error)
}

// SessionAuth obtiene la identidad y el rol exclusivamente de /auth/validate.
// Las cabeceras de identidad o rol enviadas por el cliente se descartan, de modo
// que un acceso directo al servicio no puede suplantar a un administrador.
func SessionAuth(validator SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			httpresponse.UnauthorizedGin(c, "Token de sesión no proporcionado")
			c.Abort()
			return
		}

		token := strings.TrimSpace(authHeader[len("Bearer "):])
		if token == "" {
			httpresponse.UnauthorizedGin(c, "Token de sesión no proporcionado")
			c.Abort()
			return
		}
		if validator == nil {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Servicio de autenticación no disponible")
			c.Abort()
			return
		}

		session, err := validator.ValidateToken(c.Request.Context(), token)
		if errors.Is(err, services.ErrInvalidSession) {
			httpresponse.UnauthorizedGin(c, "Sesión inválida o expirada")
			c.Abort()
			return
		}
		if err != nil {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Servicio de autenticación no disponible")
			c.Abort()
			return
		}

		userID := strings.TrimSpace(session.UserID)
		if userID == "" {
			httpresponse.UnauthorizedGin(c, "Sesión inválida o expirada")
			c.Abort()
			return
		}

		c.Request.Header.Del("X-User-ID")
		c.Request.Header.Del("X-User-Id")
		c.Request.Header.Del("X-User-Role")
		c.Request.Header.Del("X-Role")

		c.Set(authz.ContextUserIDKey, userID)
		c.Set(authz.ContextUserRoleKey, session.Role)
		c.Next()
	}
}
