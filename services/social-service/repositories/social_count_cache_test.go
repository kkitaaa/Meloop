package repositories

import (
	"context"
	"errors"
	"sync/atomic"
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

	lockKey := socialCountLockKey("user-42", "friends")
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
	lockKey := socialCountLockKey("user-42", "friends")
	token, acquired, err := cache.acquireLock(context.Background(), lockKey)
	require.NoError(t, err)
	require.True(t, acquired)
	go func() {
		time.Sleep(30 * time.Millisecond)
		_ = cache.setCount(context.Background(), key, 5, socialCountTTL)
	}()

	value, ok, err := cache.waitForCount(context.Background(), key, lockKey)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 5, value)
	require.NoError(t, cache.releaseLock(context.Background(), lockKey, token))
}

func TestSocialCountRedisCache_InvalidationFencesStalePublisher(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	defer cache.client.Close()

	ctx := context.Background()
	key := socialFriendsCountKey("user-42")
	lockKey := socialCountLockKey("user-42", "friends")
	oldToken, acquired, err := cache.acquireLock(ctx, lockKey)
	require.NoError(t, err)
	require.True(t, acquired)
	oldVersion, err := cache.getVersion(ctx, key)
	require.NoError(t, err)

	require.NoError(t, cache.invalidate(ctx, key))
	published, err := cache.publishCount(ctx, key, lockKey, oldToken, oldVersion, 3, socialCountTTL)
	require.NoError(t, err)
	require.False(t, published, "invalidation must fence a reader that captured the previous version")
	require.NoError(t, cache.releaseLock(ctx, lockKey, oldToken))

	newToken, acquired, err := cache.acquireLock(ctx, lockKey)
	require.NoError(t, err)
	require.True(t, acquired)
	newVersion, err := cache.getVersion(ctx, key)
	require.NoError(t, err)
	published, err = cache.publishCount(ctx, key, lockKey, newToken, newVersion, 4, socialCountTTL)
	require.NoError(t, err)
	require.True(t, published)
	value, ok, err := cache.getCount(ctx, key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 4, value)
}

func TestSocialCountCache_RetriesReadFencedByInvalidation(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	defer cache.client.Close()
	repo := &postgresFriendshipRepository{cache: cache}

	key := socialFriendsCountKey("changing")
	lockKey := socialCountLockKey("changing", "friends")
	var queries atomic.Int32
	staleCountCaptured := make(chan struct{})
	releaseStaleRead := make(chan struct{})
	defer func() {
		select {
		case <-releaseStaleRead:
		default:
			close(releaseStaleRead)
		}
	}()
	type result struct {
		count int
		err   error
	}
	resultCh := make(chan result, 1)
	go func() {
		count, err := repo.getSocialCount(context.Background(), key, lockKey, "friends", func(context.Context) (int, error) {
			if queries.Add(1) == 1 {
				close(staleCountCaptured)
				<-releaseStaleRead
				return 2, nil
			}
			return 3, nil
		})
		resultCh <- result{count: count, err: err}
	}()

	select {
	case <-staleCountCaptured:
	case <-time.After(time.Second):
		t.Fatal("stale count was not captured")
	}
	require.NoError(t, cache.invalidate(context.Background(), key))
	close(releaseStaleRead)

	select {
	case result := <-resultCh:
		require.NoError(t, result.err)
		require.Equal(t, 3, result.count)
	case <-time.After(time.Second):
		t.Fatal("reader did not retry after its publication was fenced")
	}
	value, ok, err := cache.getCount(context.Background(), key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 3, value)
	require.Equal(t, int32(2), queries.Load())

	count, err := repo.getSocialCount(context.Background(), key, lockKey, "friends", func(context.Context) (int, error) {
		queries.Add(1)
		return 99, nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, count, "the next request should reuse the refreshed cache value")
	require.Equal(t, int32(2), queries.Load())
}

func TestSocialCountCache_FallsBackToLoaderWhenRedisIsUnavailable(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()
	server.SetError("ERR simulated Redis outage")

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	defer cache.client.Close()
	repo := &postgresFriendshipRepository{cache: cache}

	var queries atomic.Int32
	count, err := repo.getSocialCount(context.Background(), socialFriendsCountKey("fallback"), socialCountLockKey("fallback", "friends"), "friends", func(context.Context) (int, error) {
		queries.Add(1)
		return 8, nil
	})
	require.NoError(t, err)
	require.Equal(t, 8, count)
	require.Equal(t, int32(1), queries.Load())
}

func TestSocialCountRedisCache_ExpiredLockOwnerCannotOverwriteNewOwner(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	defer cache.client.Close()

	ctx := context.Background()
	key := socialBlockedCountKey("user-42")
	lockKey := socialCountLockKey("user-42", "blocked")
	oldToken, acquired, err := cache.acquireLock(ctx, lockKey)
	require.NoError(t, err)
	require.True(t, acquired)
	version, err := cache.getVersion(ctx, key)
	require.NoError(t, err)

	server.FastForward(socialLockTTL + time.Millisecond)
	newToken, acquired, err := cache.acquireLock(ctx, lockKey)
	require.NoError(t, err)
	require.True(t, acquired)
	published, err := cache.publishCount(ctx, key, lockKey, newToken, version, 9, socialCountTTL)
	require.NoError(t, err)
	require.True(t, published)
	published, err = cache.publishCount(ctx, key, lockKey, oldToken, version, 2, socialCountTTL)
	require.NoError(t, err)
	require.False(t, published)

	value, ok, err := cache.getCount(ctx, key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 9, value)
}

func TestSocialCountCache_SlowQueryIsSharedWithoutOneSecondFallback(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	defer cache.client.Close()
	repo := &postgresFriendshipRepository{cache: cache}

	const parallelRequests = 5
	var queries atomic.Int32
	queryStarted := make(chan struct{})
	releaseQuery := make(chan struct{})
	load := func(context.Context) (int, error) {
		if queries.Add(1) == 1 {
			close(queryStarted)
			<-releaseQuery
		}
		return 17, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	type result struct {
		count int
		err   error
	}
	results := make(chan result, parallelRequests+1)
	defer func() {
		select {
		case <-releaseQuery:
		default:
			close(releaseQuery)
		}
	}()
	go func() {
		count, err := repo.getSocialCount(ctx, socialFriendsCountKey("slow"), socialCountLockKey("slow", "friends"), "friends", load)
		results <- result{count: count, err: err}
	}()
	select {
	case <-queryStarted:
	case <-time.After(time.Second):
		t.Fatal("PostgreSQL loader did not start")
	}

	ready := make(chan struct{}, parallelRequests)
	for i := 0; i < parallelRequests; i++ {
		go func() {
			ready <- struct{}{}
			count, err := repo.getSocialCount(ctx, socialFriendsCountKey("slow"), socialCountLockKey("slow", "friends"), "friends", load)
			results <- result{count: count, err: err}
		}()
	}
	for i := 0; i < parallelRequests; i++ {
		<-ready
	}
	timer := time.NewTimer(socialLockTTL + 100*time.Millisecond)
	<-timer.C
	require.Equal(t, int32(1), queries.Load(), "contenders must not query PostgreSQL while the renewed lock is held")
	lockExists, err := cache.lockExists(ctx, socialCountLockKey("slow", "friends"))
	require.NoError(t, err)
	require.True(t, lockExists, "the owner should renew its lock while PostgreSQL is still running")
	close(releaseQuery)

	for i := 0; i < parallelRequests+1; i++ {
		select {
		case result := <-results:
			require.NoError(t, result.err)
			require.Equal(t, 17, result.count)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	require.Equal(t, int32(1), queries.Load())
}

func TestSocialCountCache_FriendsAndBlockedLocksAreIndependent(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	defer cache.client.Close()
	repo := &postgresFriendshipRepository{cache: cache}

	friendsStarted := make(chan struct{})
	blockedStarted := make(chan struct{})
	release := make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	errors := make(chan error, 2)

	go func() {
		_, err := repo.getSocialCount(ctx, socialFriendsCountKey("independent"), socialCountLockKey("independent", "friends"), "friends", func(context.Context) (int, error) {
			close(friendsStarted)
			<-release
			return 1, nil
		})
		errors <- err
	}()
	go func() {
		_, err := repo.getSocialCount(ctx, socialBlockedCountKey("independent"), socialCountLockKey("independent", "blocked"), "blocked", func(context.Context) (int, error) {
			close(blockedStarted)
			<-release
			return 2, nil
		})
		errors <- err
	}()

	for name, started := range map[string]<-chan struct{}{
		"friends": friendsStarted,
		"blocked": blockedStarted,
	} {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatalf("%s count was blocked by the other count's lock", name)
		}
	}
	close(release)
	for i := 0; i < 2; i++ {
		require.NoError(t, <-errors)
	}
}

func TestRetrySocialCountInvalidationRecoversFromTransientFailure(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	cache, err := newSocialCountRedisCache("redis://" + server.Addr())
	require.NoError(t, err)
	defer cache.client.Close()

	ctx := context.Background()
	key := socialFriendsCountKey("user-42")
	require.NoError(t, cache.setCount(ctx, key, 6, socialCountTTL))
	server.SetError("ERR simulated Redis outage")
	firstAttempt := make(chan struct{})
	go func() {
		<-firstAttempt
		server.SetError("")
	}()
	attempts := 0
	err = retrySocialCountInvalidation(ctx, func(ctx context.Context) error {
		attempts++
		err := cache.invalidate(ctx, key)
		if attempts == 1 {
			close(firstAttempt)
		}
		return err
	})
	require.NoError(t, err)
	require.Equal(t, 2, attempts)
	require.False(t, socialCacheHasKey(t, cache, key), "the successful retry must clear the stale cached value")
	version, err := cache.getVersion(ctx, key)
	require.NoError(t, err)
	require.Equal(t, "1", version)
}

func TestRetrySocialCountInvalidationStopsAfterBoundedAttempts(t *testing.T) {
	attempts := 0
	err := retrySocialCountInvalidation(context.Background(), func(context.Context) error {
		attempts++
		return errors.New("Redis unavailable")
	})
	require.Error(t, err)
	require.Equal(t, socialInvalidationAttempts, attempts)
}
