package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/meloop/recommendation-service/models"
)

type userState struct {
	preferences  []string
	interactions []models.Interaction
}

type RecommendationService struct {
	mlClient *MLClient
	cache    *RecommendationCache
	mu       sync.RWMutex
	users    map[int]userState
}

func NewRecommendationService(mlClient *MLClient, cacheTTL time.Duration) *RecommendationService {
	return &RecommendationService{
		mlClient: mlClient,
		cache:    NewRecommendationCache(cacheTTL),
		users:    make(map[int]userState),
	}
}

func (service *RecommendationService) Get(ctx context.Context, userID, limit int) (models.RecommendationResponse, bool, error) {
	cacheKey := fmt.Sprintf("user:%d:%d", userID, limit)
	if response, ok := service.cache.Get(cacheKey); ok {
		return response, true, nil
	}

	service.mu.RLock()
	state := service.users[userID]
	service.mu.RUnlock()
	response, err := service.mlClient.Predict(ctx, models.RecommendationRequest{
		UserID: userID, Limit: limit, Preferences: state.preferences, Interactions: state.interactions,
	})
	if err != nil {
		return models.RecommendationResponse{}, false, err
	}
	service.cache.Set(cacheKey, response)
	return response, false, nil
}

func (service *RecommendationService) RecordInteraction(ctx context.Context, userID int, interaction models.Interaction, limit int) (models.RecommendationResponse, error) {
	if interaction.Type != "like" && interaction.Type != "friend_added" {
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
