package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/authz"
	"github.com/meloop/services/common/httpresponse"
)

const (
	ContextUserIDKey   = "gateway_user_id"
	ContextUserRoleKey = "gateway_user_role"
)

type validatedSession struct {
	Success bool `json:"success"`
	Data    struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	} `json:"data"`
}

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authenticateSession(c) {
			return
		}
		c.Next()
	}
}

// authenticateSession valida la sesión contra auth-service, elimina cualquier
// identidad o rol aportado por el cliente y publica los valores autenticados en
// el contexto. Devuelve false cuando ya respondió y abortó la cadena.
func authenticateSession(c *gin.Context) bool {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		httpresponse.UnauthorizedGin(c, "Token de sesion no proporcionado")
		c.Abort()
		return false
	}

	authURL := strings.TrimRight(os.Getenv("AUTH_SERVICE_URL"), "/")
	if authURL == "" {
		authURL = "http://localhost:8083"
	}

	request, err := http.NewRequestWithContext(
		c.Request.Context(),
		http.MethodGet,
		authURL+"/auth/validate",
		nil,
	)
	if err != nil {
		httpresponse.ErrorGin(c, http.StatusBadGateway, "BAD_GATEWAY", "No se pudo preparar la validacion de sesion")
		c.Abort()
		return false
	}
	request.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		httpresponse.ErrorGin(c, http.StatusBadGateway, "BAD_GATEWAY", "El servicio de autenticacion no esta disponible")
		c.Abort()
		return false
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		httpresponse.UnauthorizedGin(c, "Sesion invalida o expirada")
		c.Abort()
		return false
	}
	if response.StatusCode != http.StatusOK {
		httpresponse.ErrorGin(c, http.StatusBadGateway, "BAD_GATEWAY", "El servicio de autenticacion no pudo validar la sesion")
		c.Abort()
		return false
	}

	var session validatedSession
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil || !session.Success {
		httpresponse.ErrorGin(c, http.StatusBadGateway, "BAD_GATEWAY", "Respuesta invalida del servicio de autenticacion")
		c.Abort()
		return false
	}

	userID := strings.TrimSpace(session.Data.ID)
	if userID == "" {
		httpresponse.ErrorGin(c, http.StatusBadGateway, "BAD_GATEWAY", "El servicio de autenticacion no devolvio la identidad")
		c.Abort()
		return false
	}

	role := ""
	if normalized, ok := authz.NormalizeRole(session.Data.Role); ok {
		role = normalized
	}

	c.Request.Header.Del("X-User-ID")
	c.Request.Header.Del("X-User-Id")
	c.Request.Header.Del("X-User-Role")
	c.Request.Header.Del("X-Role")
	c.Request.Header.Set("X-User-ID", userID)
	if role != "" {
		c.Request.Header.Set("X-User-Role", role)
	}

	c.Set(ContextUserIDKey, userID)
	c.Set(ContextUserRoleKey, role)
	c.Set(authz.ContextUserIDKey, userID)
	c.Set(authz.ContextUserRoleKey, role)
	return true
}
