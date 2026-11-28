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

type roleProviderFunc func(ctx context.Context, userID string) (string, error)

func (f roleProviderFunc) GetRoleByID(ctx context.Context, userID string) (string, error) {
	return f(ctx, userID)
}

type mutableRoleProvider struct {
	role string
}

func (p *mutableRoleProvider) GetRoleByID(context.Context, string) (string, error) {
	return p.role, nil
}

func roleMiddlewareStatus(
	middleware gin.HandlerFunc,
	identity string,
	headers map[string]string,
) int {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if identity != "" {
			c.Set(ContextUserIDKey, identity)
		}
		c.Next()
	})
	router.Use(middleware)
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder.Code
}

func TestAdminRequiredAllowsModeratorsAndAdmins(t *testing.T) {
	for _, role := range []string{"admin", "administrator", "administrador", "moderador", "moderator"} {
		status := roleMiddlewareStatus(
			AdminRequired(testRoleProvider{role: role}),
			"user-1",
			nil,
		)
		if status != http.StatusNoContent {
			t.Errorf("role %q: status = %d, want 204", role, status)
		}
	}
}

func TestAdminRequiredRejectsOtherRoles(t *testing.T) {
	for _, role := range []string{"usuario", "root", ""} {
		status := roleMiddlewareStatus(
			AdminRequired(testRoleProvider{role: role}),
			"user-1",
			nil,
		)
		if status != http.StatusForbidden {
			t.Errorf("role %q: status = %d, want 403", role, status)
		}
	}
}

func TestAdminOnlyAllowsOnlyAdmins(t *testing.T) {
	tests := []struct {
		role string
		want int
	}{
		{role: "admin", want: http.StatusNoContent},
		{role: "administrator", want: http.StatusNoContent},
		{role: "administrador", want: http.StatusNoContent},
		{role: "moderador", want: http.StatusForbidden},
		{role: "moderator", want: http.StatusForbidden},
		{role: "usuario", want: http.StatusForbidden},
		{role: "root", want: http.StatusForbidden},
	}

	for _, tt := range tests {
		status := roleMiddlewareStatus(
			AdminOnly(testRoleProvider{role: tt.role}),
			"user-1",
			nil,
		)
		if status != tt.want {
			t.Errorf("role %q: status = %d, want %d", tt.role, status, tt.want)
		}
	}
}

func TestRoleMiddlewaresIgnoreSpoofedRoleHeaders(t *testing.T) {
	headers := map[string]string{
		"X-User-Role": "admin",
		"X-Role":      "administrator",
	}

	for _, middleware := range []gin.HandlerFunc{
		AdminRequired(testRoleProvider{role: "usuario"}),
		AdminOnly(testRoleProvider{role: "usuario"}),
		AdminOnly(testRoleProvider{role: "moderador"}),
	} {
		if status := roleMiddlewareStatus(middleware, "user-1", headers); status != http.StatusForbidden {
			t.Errorf("status = %d, want 403 pese a cabeceras suplantadas", status)
		}
	}
}

func TestRoleMiddlewaresRequireAuthenticatedIdentity(t *testing.T) {
	for _, middleware := range []gin.HandlerFunc{
		AdminRequired(testRoleProvider{role: "admin"}),
		AdminOnly(testRoleProvider{role: "admin"}),
	} {
		if status := roleMiddlewareStatus(middleware, "", nil); status != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401 sin identidad autenticada", status)
		}
	}
}

func TestRoleLookupUsesValidatedIdentityNotHeaders(t *testing.T) {
	var lookedUp string
	provider := roleProviderFunc(func(ctx context.Context, userID string) (string, error) {
		lookedUp = userID
		return "admin", nil
	})

	status := roleMiddlewareStatus(
		AdminOnly(provider),
		"verified-user",
		map[string]string{"X-User-ID": "attacker", "X-User-Role": "admin"},
	)
	if status != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", status)
	}
	if lookedUp != "verified-user" {
		t.Fatalf("identidad consultada = %q, want verified-user", lookedUp)
	}
}

func TestRoleMiddlewaresReturnUnavailableWhenLookupFails(t *testing.T) {
	provider := roleProviderFunc(func(ctx context.Context, userID string) (string, error) {
		return "", errors.New("database unavailable")
	})
	for _, middleware := range []gin.HandlerFunc{AdminRequired(provider), AdminOnly(provider)} {
		if status := roleMiddlewareStatus(middleware, "user-1", nil); status != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503", status)
		}
	}
}

func TestRoleIsRevalidatedOnEachRequest(t *testing.T) {
	provider := &mutableRoleProvider{role: "admin"}
	middleware := AdminOnly(provider)

	if status := roleMiddlewareStatus(middleware, "user-1", nil); status != http.StatusNoContent {
		t.Fatalf("status inicial = %d, want 204", status)
	}

	provider.role = "usuario"
	if status := roleMiddlewareStatus(middleware, "user-1", nil); status != http.StatusForbidden {
		t.Fatalf("status tras cambio de rol = %d, want 403", status)
	}
}
