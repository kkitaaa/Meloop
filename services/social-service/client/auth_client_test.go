package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewAuthClientUsesTimeout(t *testing.T) {
	authClient := NewAuthClient("http://auth-service")
	if authClient.httpClient.Timeout != authServiceTimeout {
		t.Fatalf("timeout = %s, want %s", authClient.httpClient.Timeout, authServiceTimeout)
	}
}

func TestAuthClientValidateToken(t *testing.T) {
	t.Run("valid session uses auth-service contract", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || r.URL.Path != "/auth/validate" {
				t.Errorf("request = %s %s, want GET /auth/validate", r.Method, r.URL.Path)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer valid-session-token" {
				t.Errorf("Authorization = %q, want Bearer valid-session-token", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":"verified-user","username":"user","email":"user@example.test"}}`))
		}))
		defer server.Close()

		userID, err := NewAuthClient(server.URL).ValidateToken(context.Background(), "valid-session-token")
		if err != nil {
			t.Fatalf("ValidateToken() error = %v", err)
		}
		if userID != "verified-user" {
			t.Fatalf("ValidateToken() = %q, want verified-user", userID)
		}
	})

	t.Run("invalid expired and revoked sessions are rejected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		for _, token := range []string{"invalid-token", "expired-token", "revoked-token"} {
			_, err := NewAuthClient(server.URL).ValidateToken(context.Background(), token)
			if !errors.Is(err, ErrInvalidSession) {
				t.Errorf("ValidateToken(%q) error = %v, want ErrInvalidSession", token, err)
			}
		}
	})

	t.Run("auth-service failures fail closed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		_, err := NewAuthClient(server.URL).ValidateToken(context.Background(), "session-token")
		if !errors.Is(err, ErrAuthServiceUnavailable) {
			t.Fatalf("ValidateToken() error = %v, want ErrAuthServiceUnavailable", err)
		}
	})

	t.Run("malformed success response fails closed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
		}))
		defer server.Close()

		_, err := NewAuthClient(server.URL).ValidateToken(context.Background(), "session-token")
		if !errors.Is(err, ErrAuthServiceUnavailable) {
			t.Fatalf("ValidateToken() error = %v, want ErrAuthServiceUnavailable", err)
		}
	})

	t.Run("request timeout fails closed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(time.Second):
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		httpClient := &http.Client{Timeout: 10 * time.Millisecond}
		_, err := NewAuthClient(server.URL, httpClient).ValidateToken(context.Background(), "session-token")
		if !errors.Is(err, ErrAuthServiceUnavailable) {
			t.Fatalf("ValidateToken() error = %v, want ErrAuthServiceUnavailable", err)
		}
	})
}
