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
	client  *redis.Client
	timeout time.Duration
}

func NewRedisFeedSnapshotStore(client *redis.Client, timeouts ...time.Duration) *RedisFeedSnapshotStore {
	timeout := 50 * time.Millisecond
	if len(timeouts) > 0 && timeouts[0] > 0 {
		timeout = timeouts[0]
	}
	return &RedisFeedSnapshotStore{client: client, timeout: timeout}
}

func (store *RedisFeedSnapshotStore) Get(ctx context.Context, key string) (models.FeedSnapshot, error) {
	queryCtx, cancel := context.WithTimeout(ctx, store.timeout)
	defer cancel()
	value, err := store.client.Get(queryCtx, key).Bytes()
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
	queryCtx, cancel := context.WithTimeout(ctx, store.timeout)
	defer cancel()
	if err := store.client.Set(queryCtx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("write feed snapshot to Redis: %w", err)
	}
	return nil
}
