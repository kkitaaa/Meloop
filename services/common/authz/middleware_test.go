package authz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

func performRequest(
	t *testing.T,
	seed func(*gin.Context),
	guard gin.HandlerFunc,
	headers map[string]string,
) (*httptest.ResponseRecorder, bool) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	if seed != nil {
		router.Use(func(c *gin.Context) {
			seed(c)
			c.Next()
		})
	}

	called := false
	router.GET("/resource", guard, func(c *gin.Context) {
		called = true
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/resource", nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	router.ServeHTTP(recorder, request)
	return recorder, called
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if recorder.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, status, recorder.Body.String())
	}

	var response httpresponse.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v; body = %s", err, recorder.Body.String())
	}
	if response.Success {
		t.Fatal("success = true, want false")
	}
	if response.Error == nil {
		t.Fatalf("error = nil, want código %q", code)
	}
	if response.Error.Code != code {
		t.Fatalf("error.code = %q, want %q", response.Error.Code, code)
	}
}

func TestRequireRolesAllowsCanonicalRoles(t *testing.T) {
	roles := []string{RoleUsuario, RoleModerador, RoleAdmin}
	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			seed := func(c *gin.Context) {
				c.Set(ContextUserIDKey, "user-1")
				c.Set(ContextUserRoleKey, role)
			}
			recorder, called := performRequest(t, seed, RequireRoles(role), map[string]string{
				"X-User-Role": "root",
				"X-Role":      "root",
			})
			if recorder.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
			}
			if !called {
				t.Fatal("el handler protegido no fue ejecutado")
			}
		})
	}
}

func TestRequireRolesNormalizesAliases(t *testing.T) {
	tests := []struct {
		name    string
		rawRole string
		allowed string
	}{
		{name: "administrador", rawRole: "administrador", allowed: RoleAdmin},
		{name: "administrator", rawRole: "ADMINISTRATOR", allowed: RoleAdmin},
		{name: "moderator", rawRole: "moderator", allowed: RoleModerador},
		{name: "moderador con espacios", rawRole: "  moderador  ", allowed: RoleModerador},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := func(c *gin.Context) {
				c.Set(ContextUserIDKey, "user-1")
				c.Set(ContextUserRoleKey, tt.rawRole)
			}
			recorder, called := performRequest(t, seed, RequireRoles(tt.allowed), nil)
			if recorder.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
			}
			if !called {
				t.Fatal("el handler protegido no fue ejecutado")
			}
		})
	}
}

func TestRequireRolesDeniesAccess(t *testing.T) {
	tests := []struct {
		name    string
		seed    func(*gin.Context)
		guard   gin.HandlerFunc
		headers map[string]string
		status  int
		code    string
	}{
		{
			name:    "identidad ausente",
			seed:    func(c *gin.Context) { c.Set(ContextUserRoleKey, RoleAdmin) },
			guard:   RequireRoles(RoleAdmin),
			headers: map[string]string{"X-User-ID": "spoofed"},
			status:  http.StatusUnauthorized,
			code:    httpresponse.ErrUnauthorized,
		},
		{
			name: "identidad vacía",
			seed: func(c *gin.Context) {
				c.Set(ContextUserIDKey, "   ")
				c.Set(ContextUserRoleKey, RoleAdmin)
			},
			guard:  RequireRoles(RoleAdmin),
			status: http.StatusUnauthorized,
			code:   httpresponse.ErrUnauthorized,
		},
		{
			name: "identidad no cadena",
			seed: func(c *gin.Context) {
				c.Set(ContextUserIDKey, 123)
				c.Set(ContextUserRoleKey, RoleAdmin)
			},
			guard:  RequireRoles(RoleAdmin),
			status: http.StatusUnauthorized,
			code:   httpresponse.ErrUnauthorized,
		},
		{
			name:   "rol ausente",
			seed:   func(c *gin.Context) { c.Set(ContextUserIDKey, "user-1") },
			guard:  RequireRoles(RoleAdmin),
			status: http.StatusForbidden,
			code:   httpresponse.ErrForbidden,
		},
		{
			name: "rol no cadena",
			seed: func(c *gin.Context) {
				c.Set(ContextUserIDKey, "user-1")
				c.Set(ContextUserRoleKey, 123)
			},
			guard:  RequireRoles(RoleAdmin),
			status: http.StatusForbidden,
			code:   httpresponse.ErrForbidden,
		},
		{
			name: "rol desconocido",
			seed: func(c *gin.Context) {
				c.Set(ContextUserIDKey, "user-1")
				c.Set(ContextUserRoleKey, "root")
			},
			guard:  RequireRoles(RoleAdmin),
			status: http.StatusForbidden,
			code:   httpresponse.ErrForbidden,
		},
		{
			name: "permisos insuficientes",
			seed: func(c *gin.Context) {
				c.Set(ContextUserIDKey, "user-1")
				c.Set(ContextUserRoleKey, RoleUsuario)
			},
			guard:  RequireRoles(RoleAdmin),
			status: http.StatusForbidden,
			code:   httpresponse.ErrForbidden,
		},
		{
			name: "sin roles permitidos",
			seed: func(c *gin.Context) {
				c.Set(ContextUserIDKey, "user-1")
				c.Set(ContextUserRoleKey, RoleAdmin)
			},
			guard:  RequireRoles(),
			status: http.StatusForbidden,
			code:   httpresponse.ErrForbidden,
		},
		{
			name: "cabeceras no otorgan permisos",
			seed: func(c *gin.Context) {
				c.Set(ContextUserIDKey, "user-1")
				c.Set(ContextUserRoleKey, RoleUsuario)
			},
			guard: RequireRoles(RoleAdmin),
			headers: map[string]string{
				"X-User-Role": RoleAdmin,
				"X-Role":      RoleAdmin,
			},
			status: http.StatusForbidden,
			code:   httpresponse.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder, called := performRequest(t, tt.seed, tt.guard, tt.headers)
			assertErrorResponse(t, recorder, tt.status, tt.code)
			if called {
				t.Fatal("el handler protegido no debió ejecutarse")
			}
		})
	}
}
