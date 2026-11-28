package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func ginSetTestMode(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)
}

type roleAuthStub struct {
	mu   sync.RWMutex
	role string
}

func newRoleAuthStub(role string) *roleAuthStub {
	return &roleAuthStub{role: role}
}

func (s *roleAuthStub) setRole(role string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.role = role
}

func (s *roleAuthStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/auth/validate" {
		http.Error(w, "unexpected path", http.StatusNotFound)
		return
	}
	if r.Header.Get("Authorization") != "Bearer valid-token" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	s.mu.RLock()
	role := s.role
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(
		w,
		`{"success":true,"data":{"id":"usr-123","username":"ana","email":"ana@example.com","role":%q},"error":null}`,
		role,
	)
}

func newGamificationStub(seen chan<- http.Header) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen != nil {
			seen <- r.Header.Clone()
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"ok":true}}`))
	}))
}

func setupGamificationGateway(
	t *testing.T,
	auth http.Handler,
	gamification *httptest.Server,
) *httptest.Server {
	t.Helper()
	authServer := httptest.NewServer(auth)
	t.Cleanup(authServer.Close)

	t.Setenv("AUTH_SERVICE_URL", authServer.URL)
	t.Setenv("GAMIFICATION_SERVICE_URL", gamification.URL)

	gateway := httptest.NewServer(funcRouter())
	t.Cleanup(gateway.Close)
	return gateway
}

func adminRequest(t *testing.T, gatewayURL string, headers map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, gatewayURL+"/api/v1/gamification/admin/rewards", nil)
	require.NoError(t, err)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestGamificationAdminAllowsValidAdminAndReplacesClientHeaders(t *testing.T) {
	ginSetTestMode(t)
	seen := make(chan http.Header, 1)
	gamification := newGamificationStub(seen)
	defer gamification.Close()

	gateway := setupGamificationGateway(t, newRoleAuthStub("admin"), gamification)
	resp := adminRequest(t, gateway.URL, map[string]string{
		"Authorization": "Bearer valid-token",
		"X-User-ID":     "spoofed-user",
		"X-User-Role":   "usuario",
		"X-Role":        "usuario",
	})

	require.Equal(t, http.StatusOK, resp.StatusCode)

	headers := <-seen
	require.Equal(t, "usr-123", headers.Get("X-User-ID"))
	require.Equal(t, "admin", headers.Get("X-User-Role"))
	require.Empty(t, headers.Get("X-Role"))
}

func TestGamificationAdminRejectsMissingAndInvalidToken(t *testing.T) {
	ginSetTestMode(t)
	seen := make(chan http.Header, 2)
	gamification := newGamificationStub(seen)
	defer gamification.Close()

	gateway := setupGamificationGateway(t, newRoleAuthStub("admin"), gamification)

	for _, headers := range []map[string]string{
		{},
		{"Authorization": "Bearer invalid-token"},
	} {
		resp := adminRequest(t, gateway.URL, headers)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}

	require.Len(t, seen, 0, "las peticiones no autenticadas no deben llegar al servicio")
}

func TestGamificationAdminRejectsNonAdminAndSpoofedHeaders(t *testing.T) {
	ginSetTestMode(t)
	tests := []struct {
		name string
		role string
	}{
		{name: "usuario", role: "usuario"},
		{name: "desconocido", role: "root"},
		{name: "sin rol", role: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen := make(chan http.Header, 1)
			gamification := newGamificationStub(seen)
			defer gamification.Close()

			gateway := setupGamificationGateway(t, newRoleAuthStub(tt.role), gamification)
			resp := adminRequest(t, gateway.URL, map[string]string{
				"Authorization": "Bearer valid-token",
				"X-User-ID":     "spoofed-user",
				"X-User-Role":   "admin",
				"X-Role":        "administrator",
			})

			require.Equal(t, http.StatusForbidden, resp.StatusCode)
			require.Len(t, seen, 0, "un rol no administrativo nunca debe llegar al servicio")
		})
	}
}

func TestGamificationAdminReflectsRoleChangeDuringActiveSession(t *testing.T) {
	ginSetTestMode(t)
	seen := make(chan http.Header, 3)
	gamification := newGamificationStub(seen)
	defer gamification.Close()

	auth := newRoleAuthStub("moderador")
	gateway := setupGamificationGateway(t, auth, gamification)
	headers := map[string]string{"Authorization": "Bearer valid-token"}

	resp := adminRequest(t, gateway.URL, headers)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	auth.setRole("admin")
	resp = adminRequest(t, gateway.URL, headers)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "admin", (<-seen).Get("X-User-Role"))

	auth.setRole("usuario")
	resp = adminRequest(t, gateway.URL, headers)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGamificationNonAdminRoutesRemainUnprotected(t *testing.T) {
	ginSetTestMode(t)
	seen := make(chan http.Header, 2)
	gamification := newGamificationStub(seen)
	defer gamification.Close()

	gateway := setupGamificationGateway(t, newRoleAuthStub("admin"), gamification)

	for _, path := range []string{
		"/api/v1/gamification/xp",
		"/api/v1/gamification/me/progress",
	} {
		resp, err := http.Post(gateway.URL+path, "application/json", nil)
		require.NoError(t, err)
		_ = resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, "path = %s", path)
	}

	require.Len(t, seen, 2)
}
