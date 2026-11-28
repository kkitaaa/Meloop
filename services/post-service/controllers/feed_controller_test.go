package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type feedIdentityVerifierStub struct {
	userID string
	err    error
}

func (verifier feedIdentityVerifierStub) AuthenticatedUserID(context.Context, string) (string, error) {
	return verifier.userID, verifier.err
}

func TestFeedControllerRejectsXUserIDThatDoesNotMatchSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewFeedController(nil, feedIdentityVerifierStub{userID: "authenticated-user"}, time.Second)
	router := gin.New()
	router.GET("/v1/feed", controller.GetMainFeed)
	request := httptest.NewRequest(http.MethodGet, "/v1/feed", nil)
	request.Header.Set("Authorization", "Bearer valid-session")
	request.Header.Set("X-User-ID", "other-user")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestFeedControllerRequiresSessionAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewFeedController(nil, feedIdentityVerifierStub{err: errors.New("auth-service unavailable")}, time.Second)
	router := gin.New()
	router.GET("/v1/feed", controller.GetMainFeed)
	request := httptest.NewRequest(http.MethodGet, "/v1/feed", nil)
	request.Header.Set("Authorization", "Bearer valid-session")
	request.Header.Set("X-User-ID", "user-42")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func TestFeedControllerRejectsMissingSessionToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewFeedController(nil, feedIdentityVerifierStub{userID: "user-42"}, time.Second)
	router := gin.New()
	router.GET("/v1/feed", controller.GetMainFeed)
	request := httptest.NewRequest(http.MethodGet, "/v1/feed", nil)
	request.Header.Set("X-User-ID", "user-42")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
