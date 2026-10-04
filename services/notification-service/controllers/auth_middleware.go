package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

const authValidationUnavailableMessage = "No se pudo validar la sesión"

func SessionRequired(authServiceURL string) gin.HandlerFunc {
	client := &http.Client{Timeout: 3 * time.Second}
	validationURL := strings.TrimRight(authServiceURL, "/") + "/auth/validate"

	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if strings.TrimSpace(authorization) == "" {
			httpresponse.UnauthorizedGin(c, "Token de sesión no proporcionado")
			c.Abort()
			return
		}

		request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, validationURL, nil)
		if err != nil {
			httpresponse.ErrorGin(c, http.StatusInternalServerError, "AUTH_CONFIGURATION_ERROR", authValidationUnavailableMessage)
			c.Abort()
			return
		}
		request.Header.Set("Authorization", authorization)

		response, err := client.Do(request)
		if err != nil {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "AUTH_SERVICE_UNAVAILABLE", authValidationUnavailableMessage)
			c.Abort()
			return
		}
		defer response.Body.Close()

		if response.StatusCode >= http.StatusInternalServerError {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "AUTH_SERVICE_UNAVAILABLE", authValidationUnavailableMessage)
			c.Abort()
			return
		}
		if response.StatusCode != http.StatusOK {
			httpresponse.UnauthorizedGin(c, "Sesión inválida o expirada")
			c.Abort()
			return
		}

		var validation struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.NewDecoder(response.Body).Decode(&validation); err != nil || !validation.Success || validation.Data.ID == "" {
			httpresponse.UnauthorizedGin(c, "Sesión inválida o expirada")
			c.Abort()
			return
		}

		c.Set("user_id", validation.Data.ID)
		c.Next()
	}
}
