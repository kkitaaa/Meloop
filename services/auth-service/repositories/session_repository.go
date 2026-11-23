package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/meloop/auth-service/models"
	"github.com/redis/go-redis/v9"
)

// SessionRepository defines operations for temporal session storage in Redis
type SessionRepository interface {
	Create(ctx context.Context, token string, user *models.SessionUser, ttl time.Duration) error
	Get(ctx context.Context, token string) (*models.SessionUser, error)
	Renew(ctx context.Context, token, userID string, ttl time.Duration) (bool, error)
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
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
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, key, string(data), ttl)
	if user != nil && user.ID != "" {
		userSessionsKey := "user_sessions:" + user.ID
		pipe.SAdd(ctx, userSessionsKey, token)
		pipe.Expire(ctx, userSessionsKey, ttl)
	}
	_, err = pipe.Exec(ctx)
	return err
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

func (r *redisSessionRepository) Renew(
	ctx context.Context,
	token string,
	userID string,
	ttl time.Duration,
) (bool, error) {
	if r.client == nil {
		return false, errors.New("redis client is not initialized")
	}
	if token == "" || userID == "" || ttl <= 0 {
		return false, nil
	}

	const renewScript = `
if redis.call("EXISTS", KEYS[1]) == 0 then
  return 0
end
redis.call("PEXPIRE", KEYS[1], ARGV[1])
redis.call("SADD", KEYS[2], ARGV[2])
redis.call("PEXPIRE", KEYS[2], ARGV[1])
return 1
`
	renewed, err := r.client.Eval(
		ctx,
		renewScript,
		[]string{"session:" + token, "user_sessions:" + userID},
		ttl.Milliseconds(),
		token,
	).Int()
	if err != nil {
		return false, err
	}
	return renewed == 1, nil
}

func (r *redisSessionRepository) Delete(ctx context.Context, token string) error {
	if r.client == nil {
		return errors.New("redis client is not initialized")
	}
	key := "session:" + token
	val, err := r.client.Get(ctx, key).Result()
	if err == nil {
		var user models.SessionUser
		if err := json.Unmarshal([]byte(val), &user); err == nil && user.ID != "" {
			_ = r.client.SRem(ctx, "user_sessions:"+user.ID, token).Err()
		}
	}
	return r.client.Del(ctx, key).Err()
}

func (r *redisSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if r.client == nil {
		return errors.New("redis client is not initialized")
	}
	if userID == "" {
		return nil
	}
	userSessionsKey := "user_sessions:" + userID
	tokens, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := r.client.Pipeline()
	for _, token := range tokens {
		pipe.Del(ctx, "session:"+token)
	}
	pipe.Del(ctx, userSessionsKey)
	_, err = pipe.Exec(ctx)
	return err
}
