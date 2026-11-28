package controllers

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

func InternalServiceRequired(expectedToken string) gin.HandlerFunc {
	expectedToken = strings.TrimSpace(expectedToken)
	return func(c *gin.Context) {
		if expectedToken == "" {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "INTERNAL_AUTH_NOT_CONFIGURED", "La autenticación entre servicios no está configurada")
			c.Abort()
			return
		}

		parts := strings.Fields(strings.TrimSpace(c.GetHeader("Authorization")))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") ||
			len(parts[1]) != len(expectedToken) ||
			subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expectedToken)) != 1 {
			httpresponse.UnauthorizedGin(c, "Autenticación de servicio inválida")
			c.Abort()
			return
		}
		c.Next()
	}
}
