package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/meloop/recommendation-service/models"
	"github.com/redis/go-redis/v9"
)

type RecommendationStore interface {
	Get(ctx context.Context, key string) (models.RecommendationResponse, bool, error)
	Set(ctx context.Context, key string, response models.RecommendationResponse) error
}

type RedisRecommendationStore struct {
	client *redis.Client
}

func NewRedisRecommendationStore(redisURL string) (*RedisRecommendationStore, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}
	options.DialTimeout = time.Second
	options.ReadTimeout = time.Second
	options.WriteTimeout = time.Second
	return &RedisRecommendationStore{client: redis.NewClient(options)}, nil
}

func (store *RedisRecommendationStore) Get(ctx context.Context, key string) (models.RecommendationResponse, bool, error) {
	value, err := store.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.RecommendationResponse{}, false, nil
	}
	if err != nil {
		return models.RecommendationResponse{}, false, fmt.Errorf("read recommendation backup from Redis: %w", err)
	}

	var response models.RecommendationResponse
	if err := json.Unmarshal(value, &response); err != nil {
		return models.RecommendationResponse{}, false, fmt.Errorf("decode recommendation backup from Redis: %w", err)
	}
	return response, true, nil
}

func (store *RedisRecommendationStore) Set(ctx context.Context, key string, response models.RecommendationResponse) error {
	value, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encode recommendation backup for Redis: %w", err)
	}
	if err := store.client.Set(ctx, key, value, 0).Err(); err != nil {
		return fmt.Errorf("write recommendation backup to Redis: %w", err)
	}
	return nil
}

func (store *RedisRecommendationStore) Close() error {
	return store.client.Close()
}
