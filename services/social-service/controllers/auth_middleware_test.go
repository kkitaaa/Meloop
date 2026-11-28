package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/social-service/client"
)

type sessionValidatorStub struct {
	userID string
	err    error
	calls  int
	token  string
}

func (s *sessionValidatorStub) ValidateToken(_ context.Context, token string) (string, error) {
	s.calls++
	s.token = token
	return s.userID, s.err
}

func serveAuthRequest(validator SessionValidator, authorization, userHeader string) (*httptest.ResponseRecorder, string) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "untrusted-context-user")
		c.Next()
	})
	resolvedUserID := ""
	router.GET("/protected", AuthRequired(validator), func(c *gin.Context) {
		resolvedUserID = c.GetString(ContextUserIDKey)
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	if userHeader != "" {
		request.Header.Set("X-User-ID", userHeader)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder, resolvedUserID
}

func TestAuthRequiredRejectsMissingTokenEvenWithUserIDHeader(t *testing.T) {
	validator := &sessionValidatorStub{userID: "verified-user"}
	recorder, _ := serveAuthRequest(validator, "", "spoofed-user")
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
	if validator.calls != 0 {
		t.Fatalf("validator calls = %d, want 0", validator.calls)
	}
}

func TestAuthRequiredUsesValidatedIdentityAndIgnoresUserIDHeader(t *testing.T) {
	validator := &sessionValidatorStub{userID: "verified-user"}
	recorder, resolvedUserID := serveAuthRequest(validator, "Bearer opaque-session", "spoofed-user")
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
	if validator.token != "opaque-session" {
		t.Fatalf("validated token = %q, want opaque-session", validator.token)
	}
	if resolvedUserID != "verified-user" {
		t.Fatalf("context user ID = %q, want verified-user", resolvedUserID)
	}
}

func TestAuthRequiredRejectsInvalidExpiredAndRevokedSessions(t *testing.T) {
	for _, token := range []string{"invalid", "expired", "revoked"} {
		t.Run(token, func(t *testing.T) {
			validator := &sessionValidatorStub{err: client.ErrInvalidSession}
			recorder, _ := serveAuthRequest(validator, "Bearer "+token, "spoofed-user")
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", recorder.Code)
			}
		})
	}
}

func TestAuthRequiredFailsClosedWhenAuthServiceIsUnavailable(t *testing.T) {
	t.Run("validator error", func(t *testing.T) {
		validator := &sessionValidatorStub{err: errors.New("connection timeout")}
		recorder, _ := serveAuthRequest(validator, "Bearer opaque-session", "spoofed-user")
		if recorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", recorder.Code)
		}
	})

	t.Run("validator missing", func(t *testing.T) {
		recorder, _ := serveAuthRequest(nil, "Bearer opaque-session", "spoofed-user")
		if recorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", recorder.Code)
		}
	})
}
