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
	socialLockWait              = 5 * time.Second
	socialLockRenewInterval     = socialLockTTL / 3
	socialCountMaxAttempts      = 3
	socialInvalidationAttempts  = 3
	socialInvalidationTimeout   = 4 * time.Second
)

var errSocialCountWaitTimeout = errors.New("timed out waiting for social count calculation")

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
	script := `for i = 1, #KEYS, 2 do
		redis.call("INCR", KEYS[i])
		redis.call("DEL", KEYS[i + 1])
	end
	return 1`
	scriptKeys := make([]string, 0, len(keys)*2)
	for _, key := range keys {
		scriptKeys = append(scriptKeys, socialCountVersionKey(key), key)
	}
	if err := c.client.Eval(ctx, script, scriptKeys).Err(); err != nil {
		return fmt.Errorf("invalidate social counts in Redis: %w", err)
	}
	return nil
}

func (c *socialCountRedisCache) getVersion(ctx context.Context, countKey string) (string, error) {
	if c == nil || c.client == nil {
		return "", errors.New("Redis social-count cache is unavailable")
	}
	version, err := c.client.Get(ctx, socialCountVersionKey(countKey)).Result()
	if errors.Is(err, redis.Nil) {
		return "0", nil
	}
	if err != nil {
		return "", fmt.Errorf("read social count version from Redis: %w", err)
	}
	return version, nil
}

func (c *socialCountRedisCache) publishCount(ctx context.Context, countKey, lockKey, token, version string, value int, ttl time.Duration) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}
	script := `local version = redis.call("GET", KEYS[1]) or "0"
	if version ~= ARGV[1] then return 0 end
	if redis.call("GET", KEYS[2]) ~= ARGV[2] then return 0 end
	redis.call("SET", KEYS[3], ARGV[3], "PX", ARGV[4])
	return 1`
	published, err := c.client.Eval(
		ctx,
		script,
		[]string{socialCountVersionKey(countKey), lockKey, countKey},
		version,
		token,
		strconv.Itoa(value),
		ttl.Milliseconds(),
	).Int()
	if err != nil {
		return false, fmt.Errorf("publish social count to Redis: %w", err)
	}
	return published == 1, nil
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

func (c *socialCountRedisCache) renewLock(ctx context.Context, key, token string) (bool, error) {
	if c == nil || c.client == nil || token == "" {
		return false, nil
	}
	script := `if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("PEXPIRE", KEYS[1], ARGV[2])
	end
	return 0`
	renewed, err := c.client.Eval(ctx, script, []string{key}, token, socialLockTTL.Milliseconds()).Int()
	if err != nil {
		return false, fmt.Errorf("renew social count lock in Redis: %w", err)
	}
	return renewed == 1, nil
}

func (c *socialCountRedisCache) startLockRenewal(ctx context.Context, key, token string) func() {
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(socialLockRenewInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				if _, err := c.renewLock(ctx, key, token); err != nil {
					c.warnRedisFailure("renew count lock", err)
				}
			}
		}
	}()
	return func() {
		close(stop)
		<-done
	}
}

func (c *socialCountRedisCache) waitForCount(ctx context.Context, countKey, lockKey string) (int, bool, error) {
	if c == nil || c.client == nil {
		return 0, false, nil
	}

	timer := time.NewTimer(socialLockWait)
	defer timer.Stop()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return 0, false, ctx.Err()
		case <-timer.C:
			value, ok, err := c.getCount(ctx, countKey)
			if err != nil || ok {
				return value, ok, err
			}
			locked, err := c.lockExists(ctx, lockKey)
			if err != nil {
				return 0, false, err
			}
			if locked {
				return 0, false, errSocialCountWaitTimeout
			}
			return 0, false, nil
		case <-ticker.C:
			value, ok, err := c.getCount(ctx, countKey)
			if err != nil || ok {
				return value, ok, err
			}
			locked, err := c.lockExists(ctx, lockKey)
			if err != nil {
				return 0, false, err
			}
			if !locked {
				return 0, false, nil
			}
		}
	}
}

func (c *socialCountRedisCache) lockExists(ctx context.Context, key string) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check social count lock in Redis: %w", err)
	}
	return exists > 0, nil
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

func socialCountVersionKey(countKey string) string {
	switch {
	case strings.HasPrefix(countKey, socialFriendCountKeyPrefix):
		return "social:friends:version:" + strings.TrimPrefix(countKey, socialFriendCountKeyPrefix)
	case strings.HasPrefix(countKey, socialBlockedCountKeyPrefix):
		return "social:blocked:version:" + strings.TrimPrefix(countKey, socialBlockedCountKeyPrefix)
	default:
		return "social:count:version:" + strings.TrimSpace(countKey)
	}
}

func socialCountLockKey(userID, countType string) string {
	return "social:lock:" + countType + ":" + strings.TrimSpace(userID)
}

func retrySocialCountInvalidation(ctx context.Context, invalidate func(context.Context) error) error {
	var lastErr error
	for attempt := 0; attempt < socialInvalidationAttempts; attempt++ {
		if err := invalidate(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt == socialInvalidationAttempts-1 {
			break
		}
		delay := time.Duration(attempt+1) * 50 * time.Millisecond
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("invalidate social counts after %d attempts: %w", attempt+1, lastErr)
		case <-timer.C:
		}
	}
	return fmt.Errorf("invalidate social counts after %d attempts: %w", socialInvalidationAttempts, lastErr)
}
