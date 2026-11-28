package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type postShareRepositoryFake struct {
	senderID    string
	recipientID string
	postID      string
	err         error
	calls       int
}

func (repository *postShareRepositoryFake) CreatePostShare(_ context.Context, senderID, recipientID, postID string) (string, error) {
	repository.senderID, repository.recipientID, repository.postID = senderID, recipientID, postID
	repository.calls++
	if repository.err != nil {
		return "", repository.err
	}
	return "message-1", nil
}

func TestPostSharePersistsReferenceAndReturnsMessageID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &postShareRepositoryFake{}
	router := gin.New()
	router.POST("/internal/messages/share", NewPostShareController(repository).Create)
	request := httptest.NewRequest(http.MethodPost, "/internal/messages/share", strings.NewReader(
		`{"sender_id":"sender-1","recipient_id":"friend-1","post_id":"post-1"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}
	if repository.senderID != "sender-1" || repository.recipientID != "friend-1" || repository.postID != "post-1" {
		t.Fatalf("stored share (%q, %q, %q)", repository.senderID, repository.recipientID, repository.postID)
	}
	if !strings.Contains(response.Body.String(), `"message_id":"message-1"`) {
		t.Fatalf("response must contain the persisted message ID: %s", response.Body.String())
	}
}

func TestPostShareValidatesRequestBeforePersistence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &postShareRepositoryFake{}
	router := gin.New()
	router.POST("/internal/messages/share", NewPostShareController(repository).Create)
	request := httptest.NewRequest(http.MethodPost, "/internal/messages/share", strings.NewReader(
		`{"sender_id":"same-user","recipient_id":"same-user","post_id":"post-1"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if repository.calls != 0 {
		t.Fatal("invalid request must not be persisted")
	}
}

func TestPostShareSurfacesPersistenceFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &postShareRepositoryFake{err: errors.New("database unavailable")}
	router := gin.New()
	router.POST("/internal/messages/share", NewPostShareController(repository).Create)
	request := httptest.NewRequest(http.MethodPost, "/internal/messages/share", strings.NewReader(
		`{"sender_id":"sender-1","recipient_id":"friend-1","post_id":"post-1"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}

func TestInternalServiceRequiredRejectsMissingAndInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/internal", InternalServiceRequired("service-secret"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	for _, authorization := range []string{"", "Bearer wrong", "Basic service-secret"} {
		request := httptest.NewRequest(http.MethodPost, "/internal", nil)
		request.Header.Set("Authorization", authorization)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("authorization %q: status = %d, want %d", authorization, response.Code, http.StatusUnauthorized)
		}
	}
	request := httptest.NewRequest(http.MethodPost, "/internal", nil)
	request.Header.Set("Authorization", "Bearer service-secret")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("valid service token status = %d, want %d", response.Code, http.StatusNoContent)
	}
}
