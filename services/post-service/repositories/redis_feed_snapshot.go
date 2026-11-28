package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/meloop/post-service/models"
	"github.com/redis/go-redis/v9"
)

type FeedSnapshotStore interface {
	Get(ctx context.Context, key string) (models.FeedSnapshot, error)
	Set(ctx context.Context, key string, snapshot models.FeedSnapshot, ttl time.Duration) error
}

type RedisFeedSnapshotStore struct {
	client *redis.Client
}

func NewRedisFeedSnapshotStore(client *redis.Client) *RedisFeedSnapshotStore {
	return &RedisFeedSnapshotStore{client: client}
}

func (store *RedisFeedSnapshotStore) Get(ctx context.Context, key string) (models.FeedSnapshot, error) {
	value, err := store.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.FeedSnapshot{}, redis.Nil
	}
	if err != nil {
		return models.FeedSnapshot{}, fmt.Errorf("read feed snapshot from Redis: %w", err)
	}

	var snapshot models.FeedSnapshot
	if err := json.Unmarshal(value, &snapshot); err != nil {
		return models.FeedSnapshot{}, fmt.Errorf("decode feed snapshot from Redis: %w", err)
	}
	return snapshot, nil
}

func (store *RedisFeedSnapshotStore) Set(ctx context.Context, key string, snapshot models.FeedSnapshot, ttl time.Duration) error {
	value, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("encode feed snapshot for Redis: %w", err)
	}
	if err := store.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("write feed snapshot to Redis: %w", err)
	}
	return nil
}
