package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/meloop/services/common/logging"
	"github.com/redis/go-redis/v9"
)

const searchCacheTTL = 5 * time.Minute

var searchLogger = logging.New("music-service")

type SearchCache interface {
	Get(context.Context, string) (map[string]any, bool, error)
	Set(context.Context, string, map[string]any, time.Duration) error
}

type RedisSearchCache struct {
	client *redis.Client
}

func NewRedisSearchCache(redisURL string) (*RedisSearchCache, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}
	options.DialTimeout = time.Second
	options.ReadTimeout = time.Second
	options.WriteTimeout = time.Second
	options.MaxRetries = -1
	return &RedisSearchCache{client: redis.NewClient(options)}, nil
}

func (c *RedisSearchCache) Close() error {
	return c.client.Close()
}

func (c *RedisSearchCache) Get(ctx context.Context, key string) (map[string]any, bool, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read search cache: %w", err)
	}
	var result map[string]any
	if err := json.Unmarshal(value, &result); err != nil || result == nil {
		return nil, false, fmt.Errorf("decode search cache: invalid response")
	}
	return result, true, nil
}

func (c *RedisSearchCache) Set(ctx context.Context, key string, value map[string]any, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode search cache: %w", err)
	}
	if err := c.client.Set(ctx, key, encoded, ttl).Err(); err != nil {
		return fmt.Errorf("write search cache: %w", err)
	}
	return nil
}

func searchCacheKey(query, itemType string, offset, limit int, market string) string {
	payload := fmt.Sprintf("%q|%q|%d|%d|%q", query, itemType, offset, limit, market)
	hash := sha256.Sum256([]byte(payload))
	return "music:search:v1:" + hex.EncodeToString(hash[:])
}

func (s *SpotifyCatalog) SetSearchCache(cache SearchCache) {
	s.cache = cache
}
