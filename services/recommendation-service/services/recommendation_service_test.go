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

func (profileRepositoryFake) GetPopular(context.Context, string, int) ([]models.PopularContent, error) {
	return nil, nil
}

type coldStartRepositoryFake struct {
	profile      models.UserProfile
	interactions []models.Interaction
	popular      []models.PopularContent
	popularCalls int
}

func (repository *coldStartRepositoryFake) Get(context.Context, string, string) (models.UserProfile, []models.Interaction, error) {
	return repository.profile, repository.interactions, nil
}

func (repository *coldStartRepositoryFake) GetPopular(_ context.Context, _ string, _ int) ([]models.PopularContent, error) {
	repository.popularCalls++
	return repository.popular, nil
}

type recommendationStoreFake struct {
	responses        map[string]models.RecommendationResponse
	cached           map[string]CachedRecommendation
	cacheTTL         time.Duration
	invalidatedUsers []int
}

func (store *recommendationStoreFake) Get(_ context.Context, key string) (models.RecommendationResponse, bool, error) {
	response, found := store.responses[key]
	return response, found, nil
}

func (store *recommendationStoreFake) Set(_ context.Context, key string, response models.RecommendationResponse) error {
	store.responses[key] = response
	return nil
}

func (store *recommendationStoreFake) GetCached(_ context.Context, key string) (CachedRecommendation, bool, error) {
	cached, found := store.cached[key]
	return cached, found, nil
}

func (store *recommendationStoreFake) SetCached(_ context.Context, key string, cached CachedRecommendation, ttl time.Duration) error {
	if store.cached == nil {
		store.cached = make(map[string]CachedRecommendation)
	}
	store.cached[key] = cached
	store.cacheTTL = ttl
	return nil
}

func (store *recommendationStoreFake) InvalidateCached(_ context.Context, userID int) error {
	store.invalidatedUsers = append(store.invalidatedUsers, userID)
	for _, recommendationType := range []string{"all", "music", "friends"} {
		delete(store.cached, recommendationCacheKey(userID, recommendationType))
	}
	return nil
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

func TestRecommendationServiceUsesPopularContentForColdStartWithoutPreviousSet(t *testing.T) {
	var mlCalls atomic.Int32
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mlCalls.Add(1)
		_ = json.NewEncoder(writer).Encode(models.RecommendationResponse{UserID: 42, Model: "test-model"})
	}))
	defer mlServer.Close()

	repository := &coldStartRepositoryFake{
		profile:      models.UserProfile{Genres: []string{"jazz"}},
		interactions: []models.Interaction{{Type: "like", TargetID: 9}},
		popular: []models.PopularContent{
			{Type: "song", ID: "spotify-track-1", Name: "So What", UsageCount: 8},
			{Type: "artist", ID: "spotify-artist-1", Name: "Miles Davis", UsageCount: 4},
		},
	}
	service := NewRecommendationServiceWithRepository(NewMLClient(mlServer.URL), time.Minute, repository)

	response, cached, err := service.GetByType(context.Background(), 42, "music", 10)
	if err != nil || cached {
		t.Fatalf("expected a fresh popularity response, response=%+v cached=%v error=%v", response, cached, err)
	}
	if repository.popularCalls != 1 || mlCalls.Load() != 0 {
		t.Fatalf("expected popularity query and no ML call, popularity_calls=%d ml_calls=%d", repository.popularCalls, mlCalls.Load())
	}
	if response.Model != "popularity" || len(response.Recommendations) != 2 {
		t.Fatalf("expected popular song and artist recommendations, got %+v", response)
	}
	first := response.Recommendations[0]
	if first.ItemKey != "spotify-track-1" || first.Type != "song" || first.Name != "So What" || first.Score != 1 {
		t.Fatalf("unexpected popular recommendation item: %+v", first)
	}
}

func TestRecommendationServiceRefreshesChangedColdStartInsteadOfUsingOldSet(t *testing.T) {
	repository := &coldStartRepositoryFake{
		popular: []models.PopularContent{{Type: "song", ID: "track-1", UsageCount: 5}},
	}
	store := &recommendationStoreFake{responses: map[string]models.RecommendationResponse{
		"recommendations:last:42:music": {
			UserID:          42,
			Recommendations: []models.RecommendationItem{{ItemID: 17}},
			Model:           "previous-model",
		},
	}}
	service := NewRecommendationServiceWithRepositoryAndStore(
		NewMLClient("http://127.0.0.1:1"),
		time.Minute,
		repository,
		store,
	)

	response, cached, err := service.GetByType(context.Background(), 42, "music", 10)
	if err != nil || cached {
		t.Fatalf("expected a fresh response for changed/no input snapshot, response=%+v cached=%v error=%v", response, cached, err)
	}
	if response.Model != "popularity" || repository.popularCalls != 1 {
		t.Fatalf("expected fresh popular results instead of the old set, response=%+v popularity_calls=%d", response, repository.popularCalls)
	}
}

func TestRecommendationServiceCachesFor24HoursAndRefreshesWhenPreferencesChange(t *testing.T) {
	var mlCalls atomic.Int32
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mlCalls.Add(1)
		var payload models.RecommendationRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode ML request: %v", err)
		}
		_ = json.NewEncoder(writer).Encode(models.RecommendationResponse{
			UserID:          payload.UserID,
			Recommendations: []models.RecommendationItem{{ItemID: int(mlCalls.Load())}, {ItemID: 99}},
			Model:           "test-model",
		})
	}))
	defer mlServer.Close()

	repository := &coldStartRepositoryFake{
		profile: models.UserProfile{
			Genres:  []string{"jazz"},
			Artists: []string{"Miles Davis"},
			Songs:   []string{"So What"},
		},
	}
	store := &recommendationStoreFake{responses: make(map[string]models.RecommendationResponse)}
	service := NewRecommendationServiceWithRepositoryAndStore(NewMLClient(mlServer.URL), time.Minute, repository, store)
	ctx := context.Background()

	first, cached, err := service.GetByType(ctx, 42, "music", 1)
	if err != nil || cached || len(first.Recommendations) != 1 {
		t.Fatalf("unexpected first calculation: response=%+v cached=%v error=%v", first, cached, err)
	}
	second, cached, err := service.GetByType(ctx, 42, "music", 10)
	if err != nil || !cached {
		t.Fatalf("expected Redis cache hit at a different result limit: response=%+v cached=%v error=%v", second, cached, err)
	}
	if len(second.Recommendations) != 2 || second.Recommendations[0].ItemID != first.Recommendations[0].ItemID {
		t.Fatalf("expected complete saved calculation to be served, first=%+v second=%+v", first, second)
	}
	if got := mlCalls.Load(); got != 1 {
		t.Fatalf("expected one ML calculation within 24 hours, got %d", got)
	}
	if store.cacheTTL != 24*time.Hour {
		t.Fatalf("expected a 24-hour Redis TTL, got %s", store.cacheTTL)
	}

	repository.profile.Genres = []string{"rock"}
	updated, cached, err := service.GetByType(ctx, 42, "music", 10)
	if err != nil || cached {
		t.Fatalf("expected preference change to trigger immediate calculation, response=%+v cached=%v error=%v", updated, cached, err)
	}
	if got := mlCalls.Load(); got != 2 {
		t.Fatalf("expected preferences change to cause second ML calculation, got %d", got)
	}
	if len(store.invalidatedUsers) != 1 || store.invalidatedUsers[0] != 42 {
		t.Fatalf("expected preference change to invalidate user's Redis marker, got %v", store.invalidatedUsers)
	}
}

func TestRecommendationServiceInvalidatesRedisCacheWhenInteractionIsRecorded(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(models.RecommendationResponse{UserID: 42, Model: "test-model"})
	}))
	defer mlServer.Close()

	store := &recommendationStoreFake{
		responses: make(map[string]models.RecommendationResponse),
		cached: map[string]CachedRecommendation{
			recommendationCacheKey(42, "music"): {Fingerprint: "previous"},
		},
	}
	service := NewRecommendationServiceWithRepositoryAndStore(NewMLClient(mlServer.URL), time.Minute, nil, store)

	if _, err := service.RecordInteraction(context.Background(), 42, models.Interaction{Type: "like", TargetID: 7}, 10); err != nil {
		t.Fatalf("record interaction: %v", err)
	}
	if len(store.invalidatedUsers) != 1 || store.invalidatedUsers[0] != 42 {
		t.Fatalf("expected Redis cache invalidation for user 42, got %v", store.invalidatedUsers)
	}
	if _, found := store.cached[recommendationCacheKey(42, "music")]; found {
		t.Fatal("expected old Redis recommendation cache entry to be removed")
	}
}

func TestRecommendationServiceKeepsMLForUsersWithEnoughRecentInteractions(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(models.RecommendationResponse{UserID: 42, Model: "test-model"})
	}))
	defer mlServer.Close()

	repository := &coldStartRepositoryFake{
		profile: models.UserProfile{InteractionCount: minimumPersonalSignals},
		popular: []models.PopularContent{{Type: "song", ID: "track-1", UsageCount: 5}},
	}
	service := NewRecommendationServiceWithRepository(NewMLClient(mlServer.URL), time.Minute, repository)

	response, _, err := service.GetByType(context.Background(), 42, "music", 10)
	if err != nil || response.Model != "test-model" {
		t.Fatalf("expected ML recommendations for a user with sufficient activity, response=%+v error=%v", response, err)
	}
	if repository.popularCalls != 0 {
		t.Fatalf("popularity query should not run for a user with sufficient activity, got %d calls", repository.popularCalls)
	}
}

func TestRecommendationServicePersistsRecommendationsAndUsesBackupOnMLFailure(t *testing.T) {
	var mlUnavailable atomic.Bool
	mlServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if mlUnavailable.Load() {
			http.Error(writer, "unavailable", http.StatusServiceUnavailable)
			return
		}
		var payload models.RecommendationRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode ML request: %v", err)
		}
		_ = json.NewEncoder(writer).Encode(models.RecommendationResponse{
			UserID:          payload.UserID,
			Recommendations: []models.RecommendationItem{{ItemID: 17}, {ItemID: 18}},
			Model:           "test-model",
		})
	}))
	defer mlServer.Close()

	store := &recommendationStoreFake{responses: make(map[string]models.RecommendationResponse)}
	service := NewRecommendationServiceWithRepositoryAndStore(NewMLClient(mlServer.URL), time.Minute, nil, store)
	ctx := context.Background()

	initial, _, err := service.GetByType(ctx, 42, "music", 10)
	if err != nil {
		t.Fatalf("get initial recommendations: %v", err)
	}
	if initial.CalculatedAt.IsZero() || initial.FromBackup {
		t.Fatalf("expected a timestamped ML response, got %+v", initial)
	}
	key := "recommendations:last:42:music"
	if _, found := store.responses[key]; !found {
		t.Fatalf("expected ML response stored under %q", key)
	}

	mlUnavailable.Store(true)
	service.cache.InvalidateUser(42)
	delete(store.cached, recommendationCacheKey(42, "music"))
	backup, cached, err := service.GetByType(ctx, 42, "music", 1)
	if err != nil || !cached {
		t.Fatalf("expected stored backup response: response=%+v cached=%v error=%v", backup, cached, err)
	}
	if !backup.FromBackup || backup.Model != "test-model" || len(backup.Recommendations) != 1 || backup.Recommendations[0].ItemID != 17 {
		t.Fatalf("expected limited prior recommendation marked as backup, got %+v", backup)
	}
	if !backup.CalculatedAt.Equal(initial.CalculatedAt) {
		t.Fatalf("expected original calculation time %s, got %s", initial.CalculatedAt, backup.CalculatedAt)
	}

	serialized, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("marshal backup response: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(serialized, &payload); err != nil {
		t.Fatalf("decode backup response JSON: %v", err)
	}
	if payload["from_backup"] != true || payload["calculated_at"] == nil {
		t.Fatalf("expected JSON backup flag and calculation date, got %s", serialized)
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
