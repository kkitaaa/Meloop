package controllers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/controllers"
	"github.com/meloop/gamification-service/services"
	"github.com/meloop/services/common/authz"
)

type fakeSessionValidator struct {
	session services.ValidatedSession
	err     error
}

func (f *fakeSessionValidator) ValidateToken(ctx context.Context, token string) (services.ValidatedSession, error) {
	return f.session, f.err
}

func newAdminRouter(validator controllers.SessionValidator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/api/v1/gamification/me/progress", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	admin := router.Group("/api/v1/gamification/admin")
	admin.Use(
		controllers.SessionAuth(validator),
		authz.RequireRoles(authz.RoleAdmin),
	)
	admin.GET("/rewards", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":    c.GetString(authz.ContextUserIDKey),
			"role":       c.GetString(authz.ContextUserRoleKey),
			"spoof_id":   c.GetHeader("X-User-ID"),
			"spoof_role": c.GetHeader("X-User-Role"),
			"spoof":      c.GetHeader("X-Role"),
		})
	})
	return router
}

func adminRequest(router *gin.Engine, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/gamification/admin/rewards", nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestAdminAllowsValidatedAdmin(t *testing.T) {
	validator := &fakeSessionValidator{
		session: services.ValidatedSession{UserID: "usr-123", Role: "admin"},
	}
	recorder := adminRequest(newAdminRouter(validator), map[string]string{
		"Authorization": "Bearer valid-token",
		"X-User-ID":     "spoofed",
		"X-User-Role":   "usuario",
		"X-Role":        "usuario",
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, want := range []string{
		`"user_id":"usr-123"`,
		`"role":"admin"`,
		`"spoof_id":""`,
		`"spoof_role":""`,
		`"spoof":""`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body = %s, want contener %s", body, want)
		}
	}
}

func TestAdminRejectsMissingAndInvalidToken(t *testing.T) {
	validator := &fakeSessionValidator{
		session: services.ValidatedSession{UserID: "usr-123", Role: "admin"},
	}
	router := newAdminRouter(validator)

	for _, headers := range []map[string]string{
		{},
		{"Authorization": "Bearer "},
	} {
		recorder := adminRequest(router, headers)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401; headers = %v", recorder.Code, headers)
		}
	}

	validator.err = services.ErrInvalidSession
	recorder := adminRequest(router, map[string]string{"Authorization": "Bearer expired"})
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 para sesión inválida", recorder.Code)
	}
}

func TestAdminRejectsNonAdminRoleDespiteSpoofedHeaders(t *testing.T) {
	validator := &fakeSessionValidator{
		session: services.ValidatedSession{UserID: "usr-123", Role: "usuario"},
	}
	recorder := adminRequest(newAdminRouter(validator), map[string]string{
		"Authorization": "Bearer valid-token",
		"X-User-Role":   "admin",
		"X-Role":        "administrator",
	})

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestAdminReflectsRoleChangeDuringActiveSession(t *testing.T) {
	validator := &fakeSessionValidator{
		session: services.ValidatedSession{UserID: "usr-123", Role: "admin"},
	}
	router := newAdminRouter(validator)
	headers := map[string]string{"Authorization": "Bearer valid-token"}

	if recorder := adminRequest(router, headers); recorder.Code != http.StatusOK {
		t.Fatalf("status inicial = %d, want 200", recorder.Code)
	}

	validator.session.Role = "usuario"
	if recorder := adminRequest(router, headers); recorder.Code != http.StatusForbidden {
		t.Fatalf("status tras cambio de rol = %d, want 403", recorder.Code)
	}
}

func TestAdminReportsUnavailableAuthService(t *testing.T) {
	validator := &fakeSessionValidator{err: errors.New("auth service caído")}
	recorder := adminRequest(newAdminRouter(validator), map[string]string{
		"Authorization": "Bearer valid-token",
	})

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestNonAdminRouteRemainsUnchanged(t *testing.T) {
	validator := &fakeSessionValidator{}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/gamification/me/progress", nil)

	newAdminRouter(validator).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
}
