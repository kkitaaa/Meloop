package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/meloop/recommendation-service/models"
	"github.com/meloop/recommendation-service/services"
)

func TestRecommendationHandlerReturnsMusicRecommendations(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var payload models.RecommendationRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode ML request: %v", err)
		}
		if payload.Type != "music" {
			t.Fatalf("expected music recommendation type, got %q", payload.Type)
		}
		_ = json.NewEncoder(writer).Encode(models.RecommendationResponse{
			UserID:          payload.UserID,
			Recommendations: []models.RecommendationItem{{ItemID: 11}},
			Model:           "test-model",
		})
	}))
	defer mlServer.Close()

	handler := recommendationHandler(services.NewRecommendationService(services.NewMLClient(mlServer.URL), time.Minute))
	request := httptest.NewRequest(http.MethodGet, "/recommendations/42/music?limit=3", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Recommendations-Source") != "ml" {
		t.Fatalf("expected ML source header, got %q", recorder.Header().Get("X-Recommendations-Source"))
	}
}

func TestRecommendationHandlerFallsBackWhenMLIsUnavailable(t *testing.T) {
	handler := recommendationHandler(services.NewRecommendationService(services.NewMLClient("http://127.0.0.1:1"), time.Minute))
	request := httptest.NewRequest(http.MethodGet, "/recommendations/friends/42", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Recommendations-Source") != "fallback" {
		t.Fatalf("expected fallback source header, got %q", recorder.Header().Get("X-Recommendations-Source"))
	}
	var response models.RecommendationResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Model != "fallback" || len(response.Recommendations) == 0 {
		t.Fatalf("expected safe fallback response, got %+v", response)
	}
}
