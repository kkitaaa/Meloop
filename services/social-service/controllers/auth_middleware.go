package controllers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/social-service/client"
)

const ContextUserIDKey = "user_id"

type SessionValidator interface {
	ValidateToken(ctx context.Context, token string) (string, error)
}

func AuthRequired(validator SessionValidator) gin.HandlerFunc {
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

		userID, err := validator.ValidateToken(c.Request.Context(), token)
		if errors.Is(err, client.ErrInvalidSession) {
			httpresponse.UnauthorizedGin(c, "Sesión inválida o expirada")
			c.Abort()
			return
		}
		if err != nil {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Servicio de autenticación no disponible")
			c.Abort()
			return
		}
		userID = strings.TrimSpace(userID)
		if userID == "" {
			httpresponse.UnauthorizedGin(c, "Sesión inválida o expirada")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}
