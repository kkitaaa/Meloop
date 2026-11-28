package controllers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/authz"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/services"
)

const (
	// ContextUserIDKey es la clave utilizada en el contexto de Gin para almacenar el ID del usuario
	ContextUserIDKey = "user_id"
	// ContextUserRoleKey es la clave utilizada para conservar el rol del usuario autenticado
	ContextUserRoleKey = "user_role"
)

type SessionValidator interface {
	ValidateToken(ctx context.Context, token string) (string, error)
}

type UserRoleProvider interface {
	GetRoleByID(ctx context.Context, userID string) (string, error)
}

// AuthRequired valida cada token contra Auth Service y establece la identidad devuelta por ese servicio.
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

// requireRoleFromBackend resuelve el rol vigente desde PostgreSQL usando la
// identidad validada por AuthRequired. Descarta las cabeceras de rol del
// cliente, normaliza el valor y delega la decisión en common/authz.
func requireRoleFromBackend(roleProvider UserRoleProvider, allowed ...string) gin.HandlerFunc {
	resolveRole := func(c *gin.Context) {
		userIDValue, exists := c.Get(ContextUserIDKey)
		userID, validUserID := userIDValue.(string)
		if !exists || !validUserID || strings.TrimSpace(userID) == "" {
			httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
			c.Abort()
			return
		}
		if roleProvider == nil {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Servicio de usuarios no disponible")
			c.Abort()
			return
		}

		roleValue, err := roleProvider.GetRoleByID(c.Request.Context(), strings.TrimSpace(userID))
		if err != nil {
			httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Servicio de usuarios no disponible")
			c.Abort()
			return
		}

		c.Request.Header.Del("X-User-Role")
		c.Request.Header.Del("X-Role")

		role, valid := authz.NormalizeRole(roleValue)
		if !valid {
			role = ""
		}
		c.Set(authz.ContextUserRoleKey, role)
		c.Set(ContextUserRoleKey, role)
	}

	requireRoles := authz.RequireRoles(allowed...)
	return func(c *gin.Context) {
		resolveRole(c)
		if c.IsAborted() {
			return
		}
		requireRoles(c)
	}
}

// AdminRequired permite a moderadores y administradores gestionar usuarios.
func AdminRequired(roleProvider UserRoleProvider) gin.HandlerFunc {
	return requireRoleFromBackend(roleProvider, authz.RoleModerador, authz.RoleAdmin)
}

// AdminOnly restringe la operación exclusivamente a administradores.
func AdminOnly(roleProvider UserRoleProvider) gin.HandlerFunc {
	return requireRoleFromBackend(roleProvider, authz.RoleAdmin)
}
