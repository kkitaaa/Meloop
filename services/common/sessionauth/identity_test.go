package sessionauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserIDReturnsValidatedSessionIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/validate" || r.Header.Get("Authorization") != "Bearer valid-token" {
			t.Fatalf("unexpected auth request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"user-1"}}`))
	}))
	defer server.Close()

	userID, err := UserID(context.Background(), server.URL, "Bearer valid-token")
	if err != nil || userID != "user-1" {
		t.Fatalf("UserID() = %q, %v; want user-1, nil", userID, err)
	}
}

func TestUserIDRejectsMissingAndInvalidSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	if _, err := UserID(context.Background(), server.URL, ""); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("missing token error = %v, want ErrUnauthorized", err)
	}
	if _, err := UserID(context.Background(), server.URL, "Bearer invalid"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("invalid token error = %v, want ErrUnauthorized", err)
	}
}

func TestUserIDReportsUnavailableAuthService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	if _, err := UserID(context.Background(), server.URL, "Bearer token"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unavailable error = %v, want ErrUnavailable", err)
	}
}
