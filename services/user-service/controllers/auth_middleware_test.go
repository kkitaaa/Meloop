package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/user-service/services"
)

type testSessionValidator struct {
	userID string
	err    error
}

func (v testSessionValidator) ValidateToken(context.Context, string) (string, error) {
	return v.userID, v.err
}

type testRoleProvider struct {
	role string
	err  error
}

func (p testRoleProvider) GetRoleByID(context.Context, string) (string, error) {
	return p.role, p.err
}

func TestAuthRequiredIgnoresSpoofedIdentityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRequired(testSessionValidator{userID: "verified-user"}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "attacker")
	req.Header.Set("X-User-Role", "admin")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without a Bearer token, got %d", recorder.Code)
	}
}

func TestAuthRequiredUsesIdentityFromAuthService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRequired(testSessionValidator{userID: "verified-user"}))
	router.GET("/", func(c *gin.Context) {
		if got, _ := c.Get(ContextUserIDKey); got != "verified-user" {
			t.Errorf("expected verified identity, got %v", got)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-session")
	req.Header.Set("X-User-ID", "attacker")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
}

func TestAuthRequiredRejectsInvalidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRequired(testSessionValidator{err: services.ErrInvalidSession}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-session")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid session, got %d", recorder.Code)
	}
}

func TestAdminRequiredIgnoresSpoofedRoleHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "user-1")
		c.Next()
	})
	router.Use(AdminRequired(testRoleProvider{role: "usuario"}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Role", "administrator")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for database role usuario, got %d", recorder.Code)
	}
}

func TestAdminRequiredUsesDatabaseRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "user-1")
		c.Next()
	})
	router.Use(AdminRequired(testRoleProvider{role: "moderador"}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Role", "usuario")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for database role moderator, got %d", recorder.Code)
	}
}

func TestAdminRequiredReturnsUnavailableWhenRoleLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "user-1")
		c.Next()
	})
	router.Use(AdminRequired(testRoleProvider{err: errors.New("database unavailable")}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when role lookup fails, got %d", recorder.Code)
	}
}
