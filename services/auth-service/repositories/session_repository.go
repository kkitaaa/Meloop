package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/meloop/auth-service/models"
)

// SessionRepository defines operations for temporal session storage in Redis
type SessionRepository interface {
	Create(ctx context.Context, token string, user *models.SessionUser, ttl time.Duration) error
	Get(ctx context.Context, token string) (*models.SessionUser, error)
	Delete(ctx context.Context, token string) error
}

type redisSessionRepository struct {
	client *redis.Client
}

// NewSessionRepository creates a new Redis implementation of SessionRepository
func NewSessionRepository(client *redis.Client) SessionRepository {
	return &redisSessionRepository{client: client}
}

func (r *redisSessionRepository) Create(ctx context.Context, token string, user *models.SessionUser, ttl time.Duration) error {
	if r.client == nil {
		return errors.New("redis client is not initialized")
	}
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	key := "session:" + token
	return r.client.Set(ctx, key, string(data), ttl).Err()
}

func (r *redisSessionRepository) Get(ctx context.Context, token string) (*models.SessionUser, error) {
	if r.client == nil {
		return nil, errors.New("redis client is not initialized")
	}
	key := "session:" + token
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Session not found or expired
		}
		return nil, err
	}

	var user models.SessionUser
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *redisSessionRepository) Delete(ctx context.Context, token string) error {
	if r.client == nil {
		return errors.New("redis client is not initialized")
	}
	key := "session:" + token
	return r.client.Del(ctx, key).Err()
}
