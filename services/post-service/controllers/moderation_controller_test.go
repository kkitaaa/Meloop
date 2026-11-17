package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
)

type moderationTargetServiceStub struct {
	target *repositories.ModerationTarget
	err    error
}

func (s moderationTargetServiceStub) GetModerationTarget(
	context.Context, string, string, string,
) (*repositories.ModerationTarget, error) {
	return s.target, s.err
}

func TestInternalModerationTargetRequiresServiceToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("INTERNAL_SERVICE_TOKEN", "expected-token")

	router := gin.New()
	router.GET(
		"/internal/moderation/targets/:target_type/:target_id",
		NewModerationController(moderationTargetServiceStub{}).GetTarget,
	)
	request := httptest.NewRequest(http.MethodGet, "/internal/moderation/targets/post/post-1", nil)
	request.Header.Set("X-User-ID", "reporter-1")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestInternalModerationTargetReturnsSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("INTERNAL_SERVICE_TOKEN", "expected-token")
	service := moderationTargetServiceStub{
		target: &repositories.ModerationTarget{
			AuthorID: "author-2", AuthorName: "Artist", Text: "post text",
			MediaURL: "https://example.test/image",
		},
	}
	router := gin.New()
	router.GET("/internal/moderation/targets/:target_type/:target_id", NewModerationController(service).GetTarget)
	request := httptest.NewRequest(http.MethodGet, "/internal/moderation/targets/post/post-1", nil)
	request.Header.Set("X-User-ID", "reporter-1")
	request.Header.Set("X-Internal-Service-Token", "expected-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}
	expected := `{"author_id":"author-2","author_name":"Artist","text":"post text","urls":["https://example.test/image"]}`
	if strings.TrimSpace(response.Body.String()) != expected {
		t.Fatalf("body = %s, want %s", response.Body.String(), expected)
	}
}
