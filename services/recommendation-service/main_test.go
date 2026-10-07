package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/meloop/recommendation-service/models"
	"github.com/meloop/recommendation-service/services"
)

type recommendationBackupFake struct {
	response models.RecommendationResponse
}

func (store recommendationBackupFake) Get(_ context.Context, _ string) (models.RecommendationResponse, bool, error) {
	return store.response, true, nil
}

func (recommendationBackupFake) Set(context.Context, string, models.RecommendationResponse) error {
	return nil
}

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
			ModelVersion:    "2.3.4",
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
	var response models.RecommendationResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ModelVersion != "2.3.4" {
		t.Fatalf("expected model version to be preserved, got %q", response.ModelVersion)
	}
}

func TestRecommendationHandlerServesRedisBackupAsJSON(t *testing.T) {
	calculatedAt := time.Date(2026, time.October, 5, 12, 0, 0, 0, time.UTC)
	store := recommendationBackupFake{response: models.RecommendationResponse{
		UserID:          42,
		Recommendations: []models.RecommendationItem{{ItemID: 99}},
		Model:           "previous-model",
		CalculatedAt:    calculatedAt,
	}}
	service := services.NewRecommendationServiceWithRepositoryAndStore(
		services.NewMLClient("http://127.0.0.1:1"),
		time.Minute,
		nil,
		store,
	)
	handler := recommendationHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/recommendations/42/music", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Recommendations-Source") != "backup" {
		t.Fatalf("expected backup response source, got %q", recorder.Header().Get("X-Recommendations-Source"))
	}
	var response models.RecommendationResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.FromBackup || !response.CalculatedAt.Equal(calculatedAt) {
		t.Fatalf("expected prior timestamp and backup flag, got %+v", response)
	}
}

func TestReadinessHandlerChecksMLService(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/ready" {
			t.Fatalf("expected ML readiness path, got %q", request.URL.Path)
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer mlServer.Close()

	handler := readinessHandler(services.NewMLClient(mlServer.URL))
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestReadinessHandlerReturnsUnavailableWhenMLIsNotReady(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer mlServer.Close()

	handler := readinessHandler(services.NewMLClient(mlServer.URL))
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
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
