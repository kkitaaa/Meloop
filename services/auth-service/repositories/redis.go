package repositories

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// InitRedis initializes a go-redis client using the provided connection string.
func InitRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)

	// Verify connection is active
	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, err
	}

	return rdb, nil
}
