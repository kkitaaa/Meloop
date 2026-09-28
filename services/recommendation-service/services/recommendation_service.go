package services

import (
	"context"
	"fmt"
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
	mu       sync.RWMutex
	users    map[int]userState
}

func NewRecommendationService(mlClient *MLClient, cacheTTL time.Duration) *RecommendationService {
	return NewRecommendationServiceWithRepository(mlClient, cacheTTL, nil)
}

func NewRecommendationServiceWithRepository(mlClient *MLClient, cacheTTL time.Duration, profiles repositories.UserProfileRepository) *RecommendationService {
	return &RecommendationService{
		mlClient: mlClient,
		cache:    NewRecommendationCache(cacheTTL),
		profiles: profiles,
		users:    make(map[int]userState),
	}
}

func (service *RecommendationService) Get(ctx context.Context, userID, limit int) (models.RecommendationResponse, bool, error) {
	return service.GetByType(ctx, userID, "all", limit)
}

func (service *RecommendationService) GetByType(ctx context.Context, userID int, recommendationType string, limit int) (models.RecommendationResponse, bool, error) {
	cacheKey := fmt.Sprintf("user:%d:%s:%d", userID, recommendationType, limit)
	if response, ok := service.cache.Get(cacheKey); ok {
		return response, true, nil
	}

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
	response, err := service.mlClient.Predict(ctx, models.RecommendationRequest{
		UserID: userID, Limit: limit, Type: recommendationType,
		Preferences: profilePreferences(profile, state.preferences), Profile: profile,
		Interactions: filterInteractions(interactions, recommendationType),
	})
	if err != nil {
		return fallbackResponse(userID, recommendationType, limit), false, nil
	}
	service.cache.Set(cacheKey, response)
	return response, false, nil
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
