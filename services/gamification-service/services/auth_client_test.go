package services_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meloop/gamification-service/services"
)

func TestAuthClientValidateTokenReturnsAuthenticatedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/validate" {
			t.Errorf("path = %q, want /auth/validate", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer session-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"usr-123","username":"ana","email":"ana@example.com","role":"administrador"},"error":null}`))
	}))
	defer server.Close()

	session, err := services.NewAuthClient(server.URL).ValidateToken(context.Background(), "session-token")
	if err != nil {
		t.Fatalf("ValidateToken devolvió error: %v", err)
	}
	if session.UserID != "usr-123" {
		t.Errorf("UserID = %q, want usr-123", session.UserID)
	}
	if session.Role != "admin" {
		t.Errorf("Role = %q, want admin", session.Role)
	}
}

func TestAuthClientRejectsInvalidSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := services.NewAuthClient(server.URL).ValidateToken(context.Background(), "expired")
	if !errors.Is(err, services.ErrInvalidSession) {
		t.Fatalf("error = %v, want ErrInvalidSession", err)
	}
}

func TestAuthClientDoesNotGrantRoleForUnknownValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"success":true,"data":{"id":"usr-123","role":"root"},"error":null}`)
	}))
	defer server.Close()

	session, err := services.NewAuthClient(server.URL).ValidateToken(context.Background(), "token")
	if err != nil {
		t.Fatalf("ValidateToken devolvió error: %v", err)
	}
	if session.Role != "" {
		t.Fatalf("Role = %q, want cadena vacía", session.Role)
	}
}

func TestAuthClientReportsUnavailableAuthService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	if _, err := services.NewAuthClient(server.URL).ValidateToken(context.Background(), "token"); err == nil {
		t.Fatal("se esperaba error ante respuesta no OK de auth-service")
	}
}
