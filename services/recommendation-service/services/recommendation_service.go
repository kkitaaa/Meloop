package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/meloop/recommendation-service/models"
	"github.com/meloop/recommendation-service/repositories"
)

type userState struct {
	preferences  []string
	interactions []models.Interaction
}

type RecommendationService struct {
	mlClient *MLClient
	cache    *RecommendationCache
	profiles repositories.UserProfileRepository
	popular  repositories.PopularityRepository
	store    RecommendationStore
	mu       sync.RWMutex
	users    map[int]userState
}

func NewRecommendationService(mlClient *MLClient, cacheTTL time.Duration) *RecommendationService {
	return NewRecommendationServiceWithRepositoryAndStore(mlClient, cacheTTL, nil, nil)
}

func NewRecommendationServiceWithRepository(mlClient *MLClient, cacheTTL time.Duration, profiles repositories.UserProfileRepository) *RecommendationService {
	return NewRecommendationServiceWithRepositoryAndStore(mlClient, cacheTTL, profiles, nil)
}

func NewRecommendationServiceWithRepositoryAndStore(mlClient *MLClient, cacheTTL time.Duration, profiles repositories.UserProfileRepository, store RecommendationStore) *RecommendationService {
	popular, _ := profiles.(repositories.PopularityRepository)
	return &RecommendationService{
		mlClient: mlClient,
		cache:    NewRecommendationCache(cacheTTL),
		profiles: profiles,
		popular:  popular,
		store:    store,
		users:    make(map[int]userState),
	}
}

func (service *RecommendationService) Get(ctx context.Context, userID, limit int) (models.RecommendationResponse, bool, error) {
	return service.GetByType(ctx, userID, "all", limit)
}

func (service *RecommendationService) GetByType(ctx context.Context, userID int, recommendationType string, limit int) (models.RecommendationResponse, bool, error) {
	service.mu.RLock()
	state := service.users[userID]
	service.mu.RUnlock()
	profile := models.UserProfile{}
	interactions := state.interactions
	if service.profiles != nil {
		persistedProfile, persistedInteractions, err := service.profiles.Get(ctx, strconv.Itoa(userID), recommendationType)
		if err != nil {
			return models.RecommendationResponse{}, false, fmt.Errorf("load recommendation profile: %w", err)
		}
		profile = persistedProfile
		interactions = append(persistedInteractions, interactions...)
	}
	preferences := profilePreferences(profile, state.preferences)
	filteredInteractions := filterInteractions(interactions, recommendationType)
	fingerprint, err := recommendationFingerprint(preferences, filteredInteractions, profile.InteractionCount)
	if err != nil {
		return models.RecommendationResponse{}, false, fmt.Errorf("fingerprint recommendation inputs: %w", err)
	}
	cacheKey := recommendationCacheKey(userID, recommendationType)
	localCacheKey := localRecommendationCacheKey(userID, recommendationType, fingerprint)
	if cacheStore, ok := service.store.(RecommendationCacheStore); ok {
		cached, found, cacheErr := cacheStore.GetCached(ctx, cacheKey)
		if cacheErr != nil {
			slog.Warn("recommendation_cache_read_failed", "user_id", userID, "type", recommendationType, "error", cacheErr)
		} else if found && cached.Fingerprint == fingerprint {
			service.cache.Set(localCacheKey, cached.Response)
			return limitRecommendations(cached.Response, limit), true, nil
		} else if found {
			if err := cacheStore.InvalidateCached(ctx, userID); err != nil {
				slog.Warn("recommendation_cache_invalidate_failed", "user_id", userID, "error", err)
			}
			service.cache.InvalidateUser(userID)
		}
	}
	if response, ok := service.cache.Get(localCacheKey); ok {
		return limitRecommendations(response, limit), true, nil
	}

	if recommendationType != "friends" &&
		personalSignalCount(preferences, filteredInteractions, profile.InteractionCount) < minimumPersonalSignals &&
		service.popular != nil {
		popular, err := service.popular.GetPopular(ctx, strconv.Itoa(userID), recommendationCacheLimit)
		if err != nil {
			return models.RecommendationResponse{}, false, fmt.Errorf("load popular recommendations: %w", err)
		}
		if len(popular) > 0 {
			response := popularRecommendationResponse(userID, popular)
			service.saveRecommendationCache(ctx, userID, recommendationType, fingerprint, response)
			if service.store != nil {
				if err := service.store.Set(ctx, recommendationStoreKey(userID, recommendationType), response); err != nil {
					slog.Warn("recommendation_backup_write_failed", "user_id", userID, "type", recommendationType, "error", err)
				}
			}
			return limitRecommendations(response, limit), false, nil
		}
	}

	response, err := service.mlClient.Predict(ctx, models.RecommendationRequest{
		UserID: userID, Limit: recommendationCacheLimit, Type: recommendationType,
		Preferences: preferences, Profile: profile,
		Interactions: filteredInteractions,
	})
	if err != nil {
		if service.store != nil {
			storedResponse, found, storeErr := service.store.Get(ctx, recommendationStoreKey(userID, recommendationType))
			if storeErr != nil {
				slog.Warn("recommendation_backup_read_failed", "user_id", userID, "type", recommendationType, "error", storeErr)
			} else if found {
				storedResponse.FromBackup = true
				return limitRecommendations(storedResponse, limit), true, nil
			}
		}
		return fallbackResponse(userID, recommendationType, limit), false, nil
	}
	response.CalculatedAt = time.Now().UTC()
	response.FromBackup = false
	service.saveRecommendationCache(ctx, userID, recommendationType, fingerprint, response)
	if service.store != nil {
		if err := service.store.Set(ctx, recommendationStoreKey(userID, recommendationType), response); err != nil {
			slog.Warn("recommendation_backup_write_failed", "user_id", userID, "type", recommendationType, "error", err)
		}
	}
	return limitRecommendations(response, limit), false, nil
}

const minimumPersonalSignals = 3
const recommendationCacheTTL = 24 * time.Hour
const recommendationCacheLimit = 50

type recommendationFingerprintInput struct {
	Preferences      []string
	Interactions     []models.Interaction
	InteractionCount int
}

func recommendationFingerprint(preferences []string, interactions []models.Interaction, interactionCount int) (string, error) {
	value, err := json.Marshal(recommendationFingerprintInput{
		Preferences:      preferences,
		Interactions:     interactions,
		InteractionCount: interactionCount,
	})
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(value)
	return hex.EncodeToString(hash[:]), nil
}

func recommendationCacheKey(userID int, recommendationType string) string {
	return recommendationCacheKeyForUser(strconv.Itoa(userID), recommendationType)
}

func recommendationCacheKeyForUser(userID, recommendationType string) string {
	return fmt.Sprintf("recommendations:calculated:%s:%s", userID, recommendationType)
}

func localRecommendationCacheKey(userID int, recommendationType, fingerprint string) string {
	return fmt.Sprintf("user:%d:%s:%s", userID, recommendationType, fingerprint)
}

func (service *RecommendationService) saveRecommendationCache(ctx context.Context, userID int, recommendationType, fingerprint string, response models.RecommendationResponse) {
	service.cache.Set(localRecommendationCacheKey(userID, recommendationType, fingerprint), response)
	if cacheStore, ok := service.store.(RecommendationCacheStore); ok {
		if err := cacheStore.SetCached(ctx, recommendationCacheKey(userID, recommendationType), CachedRecommendation{
			Fingerprint: fingerprint,
			Response:    response,
		}, recommendationCacheTTL); err != nil {
			slog.Warn("recommendation_cache_write_failed", "user_id", userID, "type", recommendationType, "error", err)
		}
	}
}

func personalSignalCount(preferences []string, interactions []models.Interaction, persistedInteractionCount int) int {
	interactionCount := len(interactions)
	if persistedInteractionCount > interactionCount {
		interactionCount = persistedInteractionCount
	}
	return len(preferences) + interactionCount
}

func popularRecommendationResponse(userID int, popular []models.PopularContent) models.RecommendationResponse {
	maxCount := int64(0)
	for _, item := range popular {
		if item.UsageCount > maxCount {
			maxCount = item.UsageCount
		}
	}

	recommendations := make([]models.RecommendationItem, 0, len(popular))
	for _, item := range popular {
		score := 0.0
		if maxCount > 0 {
			score = float64(item.UsageCount) / float64(maxCount)
		}
		recommendations = append(recommendations, models.RecommendationItem{
			ItemKey: item.ID,
			Type:    item.Type,
			Name:    item.Name,
			Score:   score,
			Reason:  "Popular en publicaciones y preferencias recientes",
		})
	}
	return models.RecommendationResponse{
		UserID:          userID,
		Recommendations: recommendations,
		Model:           "popularity",
		ModelVersion:    "1.0",
		CalculatedAt:    time.Now().UTC(),
	}
}

func recommendationStoreKey(userID int, recommendationType string) string {
	return recommendationStoreKeyForUser(strconv.Itoa(userID), recommendationType)
}

func recommendationStoreKeyForUser(userID, recommendationType string) string {
	return fmt.Sprintf("recommendations:last:%s:%s", userID, recommendationType)
}

func limitRecommendations(response models.RecommendationResponse, limit int) models.RecommendationResponse {
	if limit >= 0 && len(response.Recommendations) > limit {
		response.Recommendations = response.Recommendations[:limit]
	}
	return response
}

func profilePreferences(profile models.UserProfile, local []string) []string {
	preferences := make([]string, 0, len(profile.Genres)+len(profile.Artists)+len(profile.Songs)+len(local))
	for _, genre := range profile.Genres {
		preferences = append(preferences, "genre:"+genre)
	}
	for _, artist := range profile.Artists {
		preferences = append(preferences, "artist:"+artist)
	}
	for _, song := range profile.Songs {
		preferences = append(preferences, "song:"+song)
	}
	return append(preferences, local...)
}

func (service *RecommendationService) RecordInteraction(ctx context.Context, userID int, interaction models.Interaction, limit int) (models.RecommendationResponse, error) {
	if interaction.Type != "like" && interaction.Type != "friend_added" && interaction.Type != "comment" && interaction.Type != "post_interaction" {
		return models.RecommendationResponse{}, fmt.Errorf("unsupported interaction type %q", interaction.Type)
	}
	service.mu.Lock()
	state := service.users[userID]
	state.interactions = append(state.interactions, interaction)
	service.users[userID] = state
	service.mu.Unlock()
	service.cache.InvalidateUser(userID)
	if cacheStore, ok := service.store.(RecommendationCacheStore); ok {
		if err := cacheStore.InvalidateCached(ctx, userID); err != nil {
			slog.Warn("recommendation_cache_invalidate_failed", "user_id", userID, "error", err)
		}
	}

	response, _, err := service.Get(ctx, userID, limit)
	return response, err
}

func filterInteractions(interactions []models.Interaction, recommendationType string) []models.Interaction {
	if recommendationType == "all" {
		return interactions
	}
	filtered := make([]models.Interaction, 0, len(interactions))
	for _, interaction := range interactions {
		if (recommendationType == "music" && interaction.Type == "like") ||
			(recommendationType == "friends" && interaction.Type == "friend_added") {
			filtered = append(filtered, interaction)
		}
	}
	return filtered
}

func fallbackResponse(userID int, recommendationType string, limit int) models.RecommendationResponse {
	if limit > 5 {
		limit = 5
	}
	recommendations := make([]models.RecommendationItem, 0, limit)
	base := userID * 1000
	if recommendationType == "friends" {
		base += 500000
	}
	for index := 1; index <= limit; index++ {
		recommendations = append(recommendations, models.RecommendationItem{
			ItemID: base + index,
			Score:  0,
			Reason: "Sugerencia temporal mientras el motor de recomendaciones no está disponible",
		})
	}
	return models.RecommendationResponse{
		UserID:          userID,
		Recommendations: recommendations,
		Model:           "fallback",
	}
}
