package repositories

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	socialFriendCountKeyPrefix  = "social:friends:count:"
	socialBlockedCountKeyPrefix = "social:blocked:count:"
	socialCountTTL              = 10 * time.Minute
	socialLockTTL               = 3 * time.Second
	socialLockWait              = time.Second
)

type socialCountRedisCache struct {
	client *redis.Client
}

func newSocialCountRedisCache(redisURL string) (*socialCountRedisCache, error) {
	if strings.TrimSpace(redisURL) == "" {
		return nil, nil
	}
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL for social counts: %w", err)
	}
	options.DialTimeout = time.Second
	options.ReadTimeout = time.Second
	options.WriteTimeout = time.Second
	options.MinRetryBackoff = 50 * time.Millisecond
	options.MaxRetryBackoff = 250 * time.Millisecond
	return &socialCountRedisCache{client: redis.NewClient(options)}, nil
}

func (c *socialCountRedisCache) getCount(ctx context.Context, key string) (int, bool, error) {
	if c == nil || c.client == nil {
		return 0, false, nil
	}
	value, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("read social count from Redis: %w", err)
	}
	count, err := strconv.Atoi(value)
	if err != nil || count < 0 {
		if err == nil {
			err = errors.New("count is negative")
		}
		return 0, false, fmt.Errorf("decode social count from Redis: %w", err)
	}
	return count, true, nil
}

func (c *socialCountRedisCache) setCount(ctx context.Context, key string, value int, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	if err := c.client.Set(ctx, key, strconv.Itoa(value), ttl).Err(); err != nil {
		return fmt.Errorf("write social count to Redis: %w", err)
	}
	return nil
}

func (c *socialCountRedisCache) invalidate(ctx context.Context, keys ...string) error {
	if c == nil || c.client == nil || len(keys) == 0 {
		return nil
	}
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("invalidate social counts in Redis: %w", err)
	}
	return nil
}

func (c *socialCountRedisCache) acquireLock(ctx context.Context, key string) (string, bool, error) {
	if c == nil || c.client == nil {
		return "", false, nil
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", false, fmt.Errorf("generate social count lock token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	acquired, err := c.client.SetNX(ctx, key, token, socialLockTTL).Result()
	if err != nil {
		return "", false, fmt.Errorf("acquire social count lock in Redis: %w", err)
	}
	if !acquired {
		return token, false, nil
	}
	return token, true, nil
}

func (c *socialCountRedisCache) waitForCount(ctx context.Context, key string) (int, bool, error) {
	if c == nil || c.client == nil {
		return 0, false, nil
	}

	timer := time.NewTimer(socialLockWait)
	defer timer.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return 0, false, ctx.Err()
		case <-timer.C:
			return 0, false, nil
		case <-ticker.C:
			value, ok, err := c.getCount(ctx, key)
			if err != nil || ok {
				return value, ok, err
			}
		}
	}
}

func (c *socialCountRedisCache) releaseLock(ctx context.Context, key, token string) error {
	if c == nil || c.client == nil || token == "" {
		return nil
	}
	lockScript := `if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) else return 0 end`
	if err := c.client.Eval(ctx, lockScript, []string{key}, token).Err(); err != nil {
		return fmt.Errorf("release social count lock in Redis: %w", err)
	}
	return nil
}

func (c *socialCountRedisCache) warnRedisFailure(operation string, err error) {
	if err == nil {
		return
	}
	log.Printf("WARN social-count-cache %s: %v", operation, err)
}

func socialFriendsCountKey(userID string) string {
	return socialFriendCountKeyPrefix + strings.TrimSpace(userID)
}

func socialBlockedCountKey(userID string) string {
	return socialBlockedCountKeyPrefix + strings.TrimSpace(userID)
}

func socialCountLockKey(userID string) string {
	return "social:lock:" + strings.TrimSpace(userID)
}
