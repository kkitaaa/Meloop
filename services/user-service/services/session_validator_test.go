package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPSessionValidatorUsesAuthServiceIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/auth/validate" {
			t.Errorf("path = %q, want /auth/validate", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer session-token" {
			t.Errorf("Authorization = %q, want Bearer session-token", got)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true,"data":{"id":"verified-user"}}`))
	}))
	defer server.Close()

	validator := NewHTTPSessionValidator(server.URL)
	userID, err := validator.ValidateSession(context.Background(), "Bearer session-token")
	if err != nil {
		t.Fatalf("ValidateSession returned error: %v", err)
	}
	if userID != "verified-user" {
		t.Fatalf("user ID = %q, want verified-user", userID)
	}
}

func TestHTTPSessionValidatorRejectsMissingBearerToken(t *testing.T) {
	validator := NewHTTPSessionValidator("http://auth.example.test")
	if _, err := validator.ValidateSession(context.Background(), ""); err != ErrInvalidSession {
		t.Fatalf("ValidateSession error = %v, want ErrInvalidSession", err)
	}
}
