package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/user-service/services"
)

type stubSessionValidator struct {
	userID        string
	err           error
	authorization string
}

func (s *stubSessionValidator) ValidateSession(_ context.Context, authorization string) (string, error) {
	s.authorization = authorization
	return s.userID, s.err
}

func TestAuthRequiredWithSessionValidationIgnoresForwardedUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := &stubSessionValidator{userID: "verified-user"}
	router := gin.New()
	router.Use(AuthRequiredWithSessionValidation(validator))
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString(ContextUserIDKey))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer valid-session")
	request.Header.Set("X-User-ID", "spoofed-user")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != "verified-user" {
		t.Fatalf("response = %d %q, want 200 verified-user", response.Code, response.Body.String())
	}
	if validator.authorization != "Bearer valid-session" {
		t.Fatalf("validator received authorization %q", validator.authorization)
	}
}

func TestAuthRequiredWithSessionValidationRejectsInvalidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := &stubSessionValidator{err: services.ErrInvalidSession}
	router := gin.New()
	router.Use(AuthRequiredWithSessionValidation(validator))
	router.GET("/", func(c *gin.Context) {
		t.Fatal("protected handler ran for invalid session")
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer expired-session")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestAuthRequiredWithSessionValidationReportsAuthServiceFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := &stubSessionValidator{err: errors.New("connection failed")}
	router := gin.New()
	router.Use(AuthRequiredWithSessionValidation(validator))
	router.GET("/", func(c *gin.Context) {
		t.Fatal("protected handler ran while auth service was unavailable")
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer valid-session")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
}
