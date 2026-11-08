package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meloop/social-service/client"
)

func TestAuthClientValidateToken(t *testing.T) {
	t.Run("returns identity from validated response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/auth/validate" {
				t.Errorf("path = %q, want /auth/validate", r.URL.Path)
			}
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("unexpected authorization header")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":"validated-user"}}`))
		}))
		defer server.Close()

		userID, err := client.NewAuthClient(server.URL).ValidateToken(context.Background(), "test-token")
		if err != nil {
			t.Fatalf("ValidateToken() error = %v", err)
		}
		if userID != "validated-user" {
			t.Fatalf("ValidateToken() = %q, want validated-user", userID)
		}
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}))
		defer server.Close()

		_, err := client.NewAuthClient(server.URL).ValidateToken(context.Background(), "test-token")
		if err != client.ErrInvalidSession {
			t.Fatalf("ValidateToken() error = %v, want ErrInvalidSession", err)
		}
	})
}
