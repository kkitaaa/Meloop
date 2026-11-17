package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const rateLimitScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)
if redis.call('ZCARD', key) >= limit then
  return 0
end
redis.call('ZADD', key, now, ARGV[4])
redis.call('PEXPIRE', key, window)
return 1
`

type RedisRateLimiter struct {
	client redis.UniversalClient
	clock  func() time.Time
}

func NewRedisRateLimiter(client redis.UniversalClient) *RedisRateLimiter {
	return &RedisRateLimiter{client: client, clock: time.Now}
}

func (l *RedisRateLimiter) Allow(ctx context.Context, userID string) (bool, error) {
	now := l.clock().UTC()
	result, err := l.client.Eval(
		ctx,
		rateLimitScript,
		[]string{"moderation:reports:" + userID},
		now.UnixMilli(),
		int64(time.Hour/time.Millisecond),
		10,
		now.UnixNano(),
	).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
