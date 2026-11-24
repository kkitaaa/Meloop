package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSessionAuthAcceptsValidatedSessionWithoutRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/validate" {
			t.Errorf("path = %q, want /auth/validate", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"usr-123","username":"ana","email":"ana@example.com"},"error":null}`))
	}))
	defer authService.Close()
	t.Setenv("AUTH_SERVICE_URL", authService.URL)

	router := gin.New()
	router.GET("/users/me", SessionAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetHeader("X-User-ID"),
			"role":    c.GetHeader("X-User-Role"),
		})
	})

	request := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	request.Header.Set("X-User-ID", "spoofed")
	request.Header.Set("X-User-Role", "admin")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	if response.Body.String() != `{"role":"","user_id":"usr-123"}` {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestSessionAuthRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users/me", SessionAuth(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/users/me", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestSessionAuthRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer authService.Close()
	t.Setenv("AUTH_SERVICE_URL", authService.URL)

	router := gin.New()
	router.GET("/users/me", SessionAuth(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	request.Header.Set("Authorization", "Bearer invalid")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}
