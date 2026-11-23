package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meloop/user-service/services"
)

func TestAuthClientValidateToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/auth/validate" {
			t.Errorf("unexpected auth request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-session" {
			t.Errorf("expected Bearer token, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"verified-user"}}`))
	}))
	defer server.Close()

	client := services.NewAuthClient(server.URL)
	userID, err := client.ValidateToken(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if userID != "verified-user" {
		t.Fatalf("expected verified-user, got %q", userID)
	}
}

func TestAuthClientRejectsInvalidSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := services.NewAuthClient(server.URL).ValidateToken(context.Background(), "test-session")
	if err != services.ErrInvalidSession {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}
}
