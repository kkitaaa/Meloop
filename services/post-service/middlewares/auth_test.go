package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubTokenValidator struct {
	tokens map[string]string
	err    error
}

func (s *stubTokenValidator) ValidateToken(ctx context.Context, token string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	if uid, ok := s.tokens[token]; ok {
		return uid, nil
	}
	return "", ErrInvalidSession
}

func TestAuthMiddleware_RejectsSpoofedXUserIDAndEnforcesBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := &stubTokenValidator{
		tokens: map[string]string{"real-token-abc": "verified-user-42"},
	}

	r := gin.New()
	r.POST("/v1/posts/:id/like", RequireAuthMiddleware(validator), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"userID": c.GetString("userID")})
	})

	// 1. Petición solo con X-User-ID (intento de suplantación) debe rechazarse con 401
	reqSpoof, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
	reqSpoof.Header.Set("X-User-ID", "attacker-999")
	wSpoof := httptest.NewRecorder()
	r.ServeHTTP(wSpoof, reqSpoof)
	if wSpoof.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401 cuando solo se envía X-User-ID, obtenido %d", wSpoof.Code)
	}

	// 2. Token inválido debe rechazarse con 401
	reqBad, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
	reqBad.Header.Set("Authorization", "Bearer fake-token")
	wBad := httptest.NewRecorder()
	r.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401 con token inválido, obtenido %d", wBad.Code)
	}

	// 3. Token válido resuelve verified-user-42 e ignora X-User-ID
	reqValid, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
	reqValid.Header.Set("Authorization", "Bearer real-token-abc")
	reqValid.Header.Set("X-User-ID", "attacker-999")
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)
	if wValid.Code != http.StatusOK {
		t.Fatalf("esperado 200 con token válido, obtenido %d", wValid.Code)
	}
	if wValid.Body.String() != `{"userID":"verified-user-42"}` {
		t.Fatalf("identidad inesperada en contexto: %s", wValid.Body.String())
	}
}
