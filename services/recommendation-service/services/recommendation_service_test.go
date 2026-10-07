package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/meloop/recommendation-service/models"
)

type profileRepositoryFake struct{}

func (profileRepositoryFake) Get(context.Context, string, string) (models.UserProfile, []models.Interaction, error) {
	return models.UserProfile{Genres: []string{"jazz"}, Artists: []string{"Miles Davis"}, Songs: []string{"So What"}}, []models.Interaction{{Type: "comment", TargetID: 9}}, nil
}

func TestRecommendationServiceLoadsPersistedProfile(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var payload models.RecommendationRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode ML request: %v", err)
		}
		if len(payload.Profile.Genres) != 1 || payload.Profile.Genres[0] != "jazz" || len(payload.Interactions) != 1 {
			t.Fatalf("persisted profile was not packaged: %+v", payload)
		}
		_ = json.NewEncoder(writer).Encode(models.RecommendationResponse{UserID: payload.UserID, Model: "test-model"})
	}))
	defer mlServer.Close()

	service := NewRecommendationServiceWithRepository(NewMLClient(mlServer.URL), time.Minute, profileRepositoryFake{})
	if _, _, err := service.Get(context.Background(), 42, 10); err != nil {
		t.Fatalf("get recommendations: %v", err)
	}
}

func TestRecommendationServiceCachesAndRefreshesAfterInteraction(t *testing.T) {
	var calls atomic.Int32
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		var payload models.RecommendationRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode ML request: %v", err)
		}
		_ = json.NewEncoder(writer).Encode(models.RecommendationResponse{
			UserID:           payload.UserID,
			Recommendations:  []models.RecommendationItem{{ItemID: len(payload.Interactions) + 1}},
			Model:            "test-model",
			InteractionCount: len(payload.Interactions),
		})
	}))
	defer mlServer.Close()

	service := NewRecommendationService(NewMLClient(mlServer.URL), time.Minute)
	ctx := context.Background()

	initial, cached, err := service.Get(ctx, 42, 10)
	if err != nil || cached || initial.InteractionCount != 0 {
		t.Fatalf("unexpected initial response: response=%+v cached=%v error=%v", initial, cached, err)
	}
	_, cached, err = service.Get(ctx, 42, 10)
	if err != nil || !cached {
		t.Fatalf("expected cached response: cached=%v error=%v", cached, err)
	}

	updated, err := service.RecordInteraction(ctx, 42, models.Interaction{Type: "like", TargetID: 7}, 10)
	if err != nil || updated.InteractionCount != 1 {
		t.Fatalf("expected refreshed response: response=%+v error=%v", updated, err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected two ML calls, got %d", calls.Load())
	}
}

func TestMLClientRetriesOpensCircuitAndUsesFallback(t *testing.T) {
	var calls atomic.Int32
	var healthy atomic.Bool
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		if !healthy.Load() {
			http.Error(writer, "unavailable", http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(writer).Encode(models.RecommendationResponse{
			UserID: 42,
			Model:  "recovered-model",
		})
	}))
	defer mlServer.Close()

	client := NewMLClient(mlServer.URL)
	client.breaker.threshold = 2
	client.breaker.resetTimeout = time.Hour
	service := NewRecommendationService(client, time.Minute)

	for attempt := 0; attempt < 2; attempt++ {
		response, _, err := service.Get(context.Background(), 42, 10)
		if err != nil || response.Model != "fallback" {
			t.Fatalf("expected fallback after ML failure: response=%+v error=%v", response, err)
		}
	}
	if got := calls.Load(); got != 6 {
		t.Fatalf("expected three HTTP attempts per request, got %d total calls", got)
	}

	response, _, err := service.Get(context.Background(), 42, 10)
	if err != nil || response.Model != "fallback" {
		t.Fatalf("expected immediate fallback while circuit is open: response=%+v error=%v", response, err)
	}
	if got := calls.Load(); got != 6 {
		t.Fatalf("open circuit should not call ML service, got %d total calls", got)
	}

	healthy.Store(true)
	client.breaker.mu.Lock()
	client.breaker.openedAt = time.Now().Add(-time.Hour)
	client.breaker.mu.Unlock()
	response, err = client.Predict(context.Background(), models.RecommendationRequest{UserID: 42})
	if err != nil || response.Model != "recovered-model" {
		t.Fatalf("expected circuit recovery probe to succeed: response=%+v error=%v", response, err)
	}
	if got := calls.Load(); got != 7 {
		t.Fatalf("expected one recovery probe, got %d total calls", got)
	}
}

func TestMLClientHasExplicitHTTPTimeout(t *testing.T) {
	client := NewMLClient("http://ml-service")
	if client.client.Timeout != mlRequestTimeout || client.client.Timeout <= 0 {
		t.Fatalf("expected explicit positive HTTP timeout, got %s", client.client.Timeout)
	}
}
