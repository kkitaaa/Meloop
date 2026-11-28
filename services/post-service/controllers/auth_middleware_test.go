package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/authz"
)

type sessionVerifierStub struct {
	userID string
	err    error
}

func (verifier sessionVerifierStub) AuthenticatedUserID(context.Context, string) (string, error) {
	return verifier.userID, verifier.err
}

func TestAuthRequiredValidatesTokenAndSetsUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", AuthRequired(sessionVerifierStub{userID: "authenticated-user"}), func(c *gin.Context) {
		if c.GetString("userID") != "authenticated-user" {
			t.Errorf("userID = %q, want authenticated-user", c.GetString("userID"))
		}
		if c.GetString(authz.ContextUserIDKey) != "authenticated-user" {
			t.Errorf("authz user ID = %q, want authenticated-user", c.GetString(authz.ContextUserIDKey))
		}
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
}

func TestAuthRequiredRejectsMissingOrInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name          string
		authorization string
		verifier      services.FeedIdentityVerifier
	}{
		{name: "missing token", verifier: sessionVerifierStub{userID: "user-1"}},
		{name: "invalid token", authorization: "Bearer invalid", verifier: sessionVerifierStub{err: services.ErrInvalidFeedIdentity}},
		{name: "empty verified identity", authorization: "Bearer valid", verifier: sessionVerifierStub{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/protected", AuthRequired(tt.verifier), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authorization != "" {
				request.Header.Set("Authorization", tt.authorization)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
			}
		})
	}
}

func TestAuthRequiredReportsAuthServiceFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", AuthRequired(sessionVerifierStub{err: errors.New("auth service unavailable")}), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
}
