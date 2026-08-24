package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
)

// ProxyToService realiza un proxy inverso hacia un microservicio y maneja errores de comunicación
func ProxyToService(targetEnvVar string, defaultTarget string, stripPrefix string) gin.HandlerFunc {
	logger := logging.New("api-gateway")
	target := os.Getenv(targetEnvVar)
	if target == "" {
		target = defaultTarget
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		logger.Error("proxy_target_invalid", "target_env", targetEnvVar, "error", err)
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Manejo personalizado de errores cuando el microservicio de destino no responde o falla
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("proxy_request_failed", "target_env", targetEnvVar, "method", r.Method, "path", r.URL.Path, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)

		response := httpresponse.Response{
			Success: false,
			Data:    nil,
			Error: &httpresponse.ErrorBody{
				Code:    "BAD_GATEWAY",
				Message: "El microservicio de destino no está disponible o ha fallado la conexión",
				Details: err.Error(),
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}

	return func(c *gin.Context) {
		originalPath := c.Request.URL.Path

		// Recortar prefijo si está configurado
		if stripPrefix != "" && strings.HasPrefix(originalPath, stripPrefix) {
			c.Request.URL.Path = strings.TrimPrefix(originalPath, stripPrefix)
			if c.Request.URL.Path == "" {
				c.Request.URL.Path = "/"
			}
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
