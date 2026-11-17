package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPIdentityResolverUsesAuthServiceSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/auth/validate" {
			t.Errorf("path = %q, want /auth/validate", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer session-token" {
			t.Errorf("authorization header was not forwarded")
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true,"data":{"id":"user-123"}}`))
	}))
	defer server.Close()

	resolver := NewHTTPIdentityResolver(server.URL, server.Client())
	userID, err := resolver.Resolve(context.Background(), "Bearer session-token")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("user ID = %q, want user-123", userID)
	}
}

func TestHTTPIdentityResolverRejectsMissingAndInvalidSessions(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		resolver := NewHTTPIdentityResolver("http://auth.example.test", nil)
		_, err := resolver.Resolve(context.Background(), "")
		if !errors.Is(err, ErrUnauthenticated) {
			t.Fatalf("Resolve() error = %v, want ErrUnauthenticated", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		resolver := NewHTTPIdentityResolver(server.URL, server.Client())
		_, err := resolver.Resolve(context.Background(), "Bearer invalid")
		if !errors.Is(err, ErrUnauthenticated) {
			t.Fatalf("Resolve() error = %v, want ErrUnauthenticated", err)
		}
	})
}
