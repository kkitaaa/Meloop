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
