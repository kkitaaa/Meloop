package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const blockedUsersCacheTTL = 60 * time.Second

type BlockedUsersProvider interface {
	BlockedWith(ctx context.Context, userID string) ([]string, error)
}

type FreshBlockedUsersProvider interface {
	BlockedWithFresh(ctx context.Context, userID string) ([]string, error)
}

type blockedUsersSource interface {
	ListUserIDs(ctx context.Context) ([]string, error)
}

type blockedCandidatesProvider interface {
	GetBlockedCandidateIDs(ctx context.Context, userID string, candidateIDs []string) ([]string, error)
}

type blockedUsersCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

type RedisBlockedUsers struct {
	cache  blockedUsersCache
	users  blockedUsersSource
	social blockedCandidatesProvider
}

func NewRedisBlockedUsers(cache redis.Cmdable, users blockedUsersSource, social blockedCandidatesProvider) *RedisBlockedUsers {
	return newRedisBlockedUsers(redisBlockedUsersCache{client: cache}, users, social)
}

func newRedisBlockedUsers(cache blockedUsersCache, users blockedUsersSource, social blockedCandidatesProvider) *RedisBlockedUsers {
	return &RedisBlockedUsers{cache: cache, users: users, social: social}
}

func (s *RedisBlockedUsers) BlockedWith(ctx context.Context, userID string) ([]string, error) {
	if userID == "" || s.cache == nil || s.users == nil || s.social == nil {
		return nil, errors.New("blocked users provider is not configured")
	}
	key := blockedUsersCacheKey(userID)
	cached, err := s.cache.Get(ctx, key)
	if err == nil {
		var blocked []string
		if err := json.Unmarshal([]byte(cached), &blocked); err != nil {
			slog.Warn("blocked_users_cache_decode_failed", "user_id", userID, "error", err)
		} else {
			if blocked == nil {
				blocked = []string{}
			}
			return blocked, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		slog.Warn("blocked_users_cache_read_failed", "user_id", userID, "error", err)
	}

	blocked, err := s.fetchBlockedUsers(ctx, userID)
	if err != nil {
		return nil, err
	}
	s.cacheBlockedUsers(ctx, userID, blocked)
	return blocked, nil
}

func (s *RedisBlockedUsers) BlockedWithFresh(ctx context.Context, userID string) ([]string, error) {
	if userID == "" || s.cache == nil || s.users == nil || s.social == nil {
		return nil, errors.New("blocked users provider is not configured")
	}
	blocked, err := s.fetchBlockedUsers(ctx, userID)
	if err != nil {
		return nil, err
	}
	s.cacheBlockedUsers(ctx, userID, blocked)
	return blocked, nil
}

func (s *RedisBlockedUsers) fetchBlockedUsers(ctx context.Context, userID string) ([]string, error) {
	candidates, err := s.users.ListUserIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users for block lookup: %w", err)
	}
	blocked, err := s.social.GetBlockedCandidateIDs(ctx, userID, candidates)
	if err != nil {
		return nil, fmt.Errorf("get blocked candidates from social-service: %w", err)
	}
	if blocked == nil {
		blocked = []string{}
	}
	return blocked, nil
}

func (s *RedisBlockedUsers) cacheBlockedUsers(ctx context.Context, userID string, blocked []string) {
	value, err := json.Marshal(blocked)
	if err != nil {
		slog.Error("blocked_users_cache_encode_failed", "user_id", userID, "error", err)
		return
	}
	if err := s.cache.Set(ctx, blockedUsersCacheKey(userID), string(value), blockedUsersCacheTTL); err != nil {
		slog.Warn("blocked_users_cache_write_failed", "user_id", userID, "error", err)
	}
}

func blockedUsersCacheKey(userID string) string {
	digest := sha256.Sum256([]byte(userID))
	return "post-service:blocked-users:" + hex.EncodeToString(digest[:])
}

type redisBlockedUsersCache struct {
	client redis.Cmdable
}

func (c redisBlockedUsersCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c redisBlockedUsersCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}
