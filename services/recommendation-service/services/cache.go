package services

import (
	"strings"
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

func (cache *RecommendationCache) InvalidateUser(userID string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	prefix := "user:" + userID + ":"
	for key := range cache.entries {
		if strings.HasPrefix(key, prefix) {
			delete(cache.entries, key)
		}
	}
}
