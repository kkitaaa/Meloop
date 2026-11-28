package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/authz"
	"github.com/meloop/services/common/httpresponse"
)

func AuthRequired(verifier services.FeedIdentityVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(strings.TrimSpace(c.GetHeader("Authorization")))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			httpresponse.UnauthorizedGin(c, "Token de sesión no proporcionado")
			c.Abort()
			return
		}
		if verifier == nil {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Servicio de autenticación no disponible")
			c.Abort()
			return
		}

		userID, err := verifier.AuthenticatedUserID(c.Request.Context(), "Bearer "+parts[1])
		if errors.Is(err, services.ErrInvalidFeedIdentity) {
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

		c.Set("userID", userID)
		c.Set(authz.ContextUserIDKey, userID)
		c.Next()
	}
}
