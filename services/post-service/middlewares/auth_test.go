package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubTokenValidator struct {
	tokens map[string]string
}

func (s *stubTokenValidator) ValidateToken(_ context.Context, token string) (string, error) {
	if uid, ok := s.tokens[token]; ok {
		return uid, nil
	}
	return "", errors.New("invalid token")
}

func TestAuthMiddleware_RejectsSpoofedXUserIDAndEnforcesBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := &stubTokenValidator{
		tokens: map[string]string{
			"valid-token-123": "authenticated-user-1",
		},
	}

	r := gin.New()
	r.Use(OptionalAuthMiddleware(validator))
	r.POST("/api/v1/posts", RequireAuthMiddleware(validator), func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"user_id": c.GetString("userID")})
	})

	// 1. Solo X-User-ID sin Bearer -> 401
	reqSpoof, _ := http.NewRequest(http.MethodPost, "/api/v1/posts", nil)
	reqSpoof.Header.Set("X-User-ID", "attacker-id")
	wSpoof := httptest.NewRecorder()
	r.ServeHTTP(wSpoof, reqSpoof)
	if wSpoof.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401 cuando se envía solo X-User-ID, obtenido %d", wSpoof.Code)
	}

	// 2. Bearer válido + X-User-ID falso -> prevalece el usuario autenticado
	reqValid, _ := http.NewRequest(http.MethodPost, "/api/v1/posts", nil)
	reqValid.Header.Set("Authorization", "Bearer valid-token-123")
	reqValid.Header.Set("X-User-ID", "attacker-id")
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)
	if wValid.Code != http.StatusCreated {
		t.Fatalf("esperado 201 con token válido, obtenido %d", wValid.Code)
	}
	if wValid.Body.String() != `{"user_id":"authenticated-user-1"}` {
		t.Fatalf("identidad inesperada: %s", wValid.Body.String())
	}
}
