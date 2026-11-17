package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func TestSocialCountRedisCache_StoresValueAndTTL(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	require.NotNil(t, cache)

	key := socialFriendsCountKey("user-42")
	require.Equal(t, "social:friends:count:user-42", key)
	require.NoError(t, cache.setCount(context.Background(), key, 7, socialCountTTL))

	value, ok, err := cache.getCount(context.Background(), key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 7, value)

	ttl := server.TTL(key)
	require.GreaterOrEqual(t, ttl, 9*time.Minute)
	require.LessOrEqual(t, ttl, 10*time.Minute)
}

func TestSocialCountRedisCache_LockIsExclusiveAndReleased(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	require.NotNil(t, cache)

	lockKey := socialCountLockKey("user-42")
	token, acquired, err := cache.acquireLock(context.Background(), lockKey)
	require.NoError(t, err)
	require.True(t, acquired)
	require.NotEmpty(t, token)

	_, acquired, err = cache.acquireLock(context.Background(), lockKey)
	require.NoError(t, err)
	require.False(t, acquired)

	require.NoError(t, cache.releaseLock(context.Background(), lockKey, token))
	nextToken, acquired, err := cache.acquireLock(context.Background(), lockKey)
	require.NoError(t, err)
	require.True(t, acquired)

	server.Set(lockKey, "another-owner")
	require.NoError(t, cache.releaseLock(context.Background(), lockKey, nextToken))
	value, err := server.Get(lockKey)
	require.NoError(t, err)
	require.Equal(t, "another-owner", value, "a stale owner must not release another request's lock")
}

func TestSocialCountRedisCache_WaitForConcurrentCalculation(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	defer cache.client.Close()

	key := socialFriendsCountKey("user-42")
	go func() {
		time.Sleep(30 * time.Millisecond)
		_ = cache.setCount(context.Background(), key, 5, socialCountTTL)
	}()

	value, ok, err := cache.waitForCount(context.Background(), key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 5, value)
}
