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

		header := strings.TrimSpace(c.GetHeader("Authorization"))
		token, found := strings.CutPrefix(header, "Bearer ")
		if !found || len(token) != len(expectedToken) ||
			subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
			httpresponse.UnauthorizedGin(c, "Autenticación de servicio inválida")
			c.Abort()
			return
		}
		c.Next()
	}
}
