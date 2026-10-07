package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/meloop/recommendation-service/models"
)

type cacheEntry struct {
	response  models.RecommendationResponse
	expiresAt time.Time
}

type RecommendationCache struct {
	mu      sync.RWMutex
	ttl     time.Duration
	entries map[string]cacheEntry
}

func NewRecommendationCache(ttl time.Duration) *RecommendationCache {
	return &RecommendationCache{ttl: ttl, entries: make(map[string]cacheEntry)}
}

func (cache *RecommendationCache) Get(key string) (models.RecommendationResponse, bool) {
	cache.mu.RLock()
	entry, exists := cache.entries[key]
	cache.mu.RUnlock()
	if !exists || time.Now().After(entry.expiresAt) {
		if exists {
			cache.Delete(key)
		}
		return models.RecommendationResponse{}, false
	}
	return entry.response, true
}

func (cache *RecommendationCache) Set(key string, response models.RecommendationResponse) {
	cache.mu.Lock()
	cache.entries[key] = cacheEntry{response: response, expiresAt: time.Now().Add(cache.ttl)}
	cache.mu.Unlock()
}

func (cache *RecommendationCache) Delete(key string) {
	cache.mu.Lock()
	delete(cache.entries, key)
	cache.mu.Unlock()
}

func (cache *RecommendationCache) InvalidateUser(userID int) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	for key := range cache.entries {
		var cachedUserID int
		if _, err := fmt.Sscanf(key, "user:%d:", &cachedUserID); err == nil && cachedUserID == userID {
			delete(cache.entries, key)
		}
	}
}
