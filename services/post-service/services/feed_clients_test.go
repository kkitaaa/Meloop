package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestSocialFeedClientUsesCentralBlockExclusionEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/internal/blocks/exclude" {
			t.Errorf("request = %s %s, want POST /internal/blocks/exclude", request.Method, request.URL.Path)
		}
		var payload struct {
			UserID       string   `json:"user_id"`
			CandidateIDs []string `json:"candidate_ids"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode exclusion request: %v", err)
		}
		if payload.UserID != "viewer" || !reflect.DeepEqual(payload.CandidateIDs, []string{"author-a", "author-b"}) {
			t.Errorf("request payload = %+v", payload)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true,"data":{"candidate_ids":["author-a"],"excluded_ids":["author-b"]}}`))
	}))
	defer server.Close()

	client := &SocialFeedClient{baseURL: server.URL, client: server.Client()}
	allowed, err := client.ExcludeBlocked(context.Background(), "viewer", []string{"author-a", "author-b"})
	if err != nil {
		t.Fatalf("exclude blocked authors: %v", err)
	}
	if !reflect.DeepEqual(allowed, []string{"author-a"}) {
		t.Fatalf("allowed authors = %v, want [author-a]", allowed)
	}
}

func TestRecommendationProviderUsesRecommendationIDsAsPostIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/recommendations/42" || request.URL.Query().Get("limit") != "200" {
			t.Errorf("request = %s %s, want GET /recommendations/42?limit=200", request.Method, request.URL.String())
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"recommendations":[{"item_key":"post-key","score":0.8},{"item_id":17,"score":0.4}]}`))
	}))
	defer server.Close()

	client := &HTTPFeedRecommendationProvider{baseURL: server.URL, client: server.Client()}
	recommendations, err := client.GetPostRecommendations(context.Background(), "42", 200)
	if err != nil {
		t.Fatalf("get recommendations: %v", err)
	}
	want := []FeedRecommendation{{PostID: "post-key", Score: 0.8}, {PostID: "17", Score: 0.4}}
	if !reflect.DeepEqual(recommendations, want) {
		t.Fatalf("recommendations = %+v, want %+v", recommendations, want)
	}
}

func TestRecommendationProviderSkipsNonNumericUserIDs(t *testing.T) {
	client := NewHTTPFeedRecommendationProvider("http://127.0.0.1:1")
	recommendations, err := client.GetPostRecommendations(context.Background(), "user-uuid", 200)
	if err != nil {
		t.Fatalf("non-numeric user ID should skip recommendations, got error: %v", err)
	}
	if len(recommendations) != 0 {
		t.Fatalf("recommendations = %+v, want empty", recommendations)
	}
}

func TestRecommendationProviderReturnsErrorWhenServiceIsDown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := &HTTPFeedRecommendationProvider{baseURL: server.URL, client: server.Client()}
	if _, err := client.GetPostRecommendations(context.Background(), "42", 200); err == nil {
		t.Fatal("expected an error from unavailable recommendation-service")
	}
}

func TestFeedIdentityVerifierUsesAuthenticatedSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/auth/validate" || request.Header.Get("Authorization") != "Bearer valid-session" {
			t.Errorf("validation request = %s with authorization %q", request.URL.Path, request.Header.Get("Authorization"))
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true,"data":{"id":"user-42"}}`))
	}))
	defer server.Close()

	verifier := NewHTTPFeedIdentityVerifier(server.URL, time.Second)
	userID, err := verifier.AuthenticatedUserID(context.Background(), "Bearer valid-session")
	if err != nil {
		t.Fatalf("validate session: %v", err)
	}
	if userID != "user-42" {
		t.Fatalf("authenticated user ID = %q, want user-42", userID)
	}
}

func TestValidateFeedUserIDRejectsUnsafeIdentifiers(t *testing.T) {
	for _, userID := range []string{"", " leading-space", "has\nnewline", string(make([]byte, 65))} {
		if ValidateFeedUserID(userID) {
			t.Errorf("ValidateFeedUserID(%q) = true, want false", userID)
		}
	}
	for _, userID := range []string{"user-42", "550e8400-e29b-41d4-a716-446655440000"} {
		if !ValidateFeedUserID(userID) {
			t.Errorf("ValidateFeedUserID(%q) = false, want true", userID)
		}
	}
}
