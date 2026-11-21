package repositories

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/social-service/models"
	"github.com/stretchr/testify/require"
)

type socialCountQueryTracer struct {
	mu                      sync.Mutex
	friendCountQueries      int
	blockedCountQueries     int
	blockFirstFriendQuery   bool
	firstFriendQuery        chan struct{}
	releaseFirstFriendQuery chan struct{}
}

func (tracer *socialCountQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	query := strings.ToLower(data.SQL)
	tracer.mu.Lock()
	blockFirst := false
	if strings.Contains(query, "count(*)") && strings.Contains(query, "from amistad") {
		tracer.friendCountQueries++
		blockFirst = tracer.blockFirstFriendQuery && tracer.friendCountQueries == 1
	} else if strings.Contains(query, "count(*)") && strings.Contains(query, "from bloqueo") {
		tracer.blockedCountQueries++
	}
	tracer.mu.Unlock()

	if blockFirst {
		close(tracer.firstFriendQuery)
		<-tracer.releaseFirstFriendQuery
	}
	return ctx
}

func (*socialCountQueryTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (tracer *socialCountQueryTracer) queryCounts() (friends, blocked int) {
	tracer.mu.Lock()
	defer tracer.mu.Unlock()
	return tracer.friendCountQueries, tracer.blockedCountQueries
}

func TestSocialCountCacheAsideInvalidationFallbackAndConcurrency(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured; integration test requires an isolated test database")
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(poolConfig.ConnConfig.Database), "test",
		"TEST_DATABASE_URL must point to a database whose name contains 'test'")
	tracer := &socialCountQueryTracer{
		blockFirstFriendQuery:   true,
		firstFriendQuery:        make(chan struct{}),
		releaseFirstFriendQuery: make(chan struct{}),
	}
	poolConfig.ConnConfig.Tracer = tracer
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(context.Background()))
	defer pool.Close()

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	user1 := "social-count-" + runID + "-1"
	user2 := "social-count-" + runID + "-2"
	for index, userID := range []string{user1, user2} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO usuario (id_usuario, username, correo, contrasena_hash)
			VALUES ($1, $2, $3, $4)`,
			userID,
			fmt.Sprintf("socialcount%s%d", runID, index),
			fmt.Sprintf("socialcount%s%d@example.test", runID, index),
			"test-password-hash",
		)
		require.NoError(t, err)
	}
	defer func() {
		_, cleanupErr := pool.Exec(context.Background(), `DELETE FROM usuario WHERE id_usuario = ANY($1)`, []string{user1, user2})
		require.NoError(t, cleanupErr)
	}()

	var redisServer *miniredis.Miniredis
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisServer, err = miniredis.Run()
		require.NoError(t, err)
		redisURL = "redis://" + redisServer.Addr()
		defer redisServer.Close()
	}
	cache, err := newSocialCountRedisCache(redisURL)
	require.NoError(t, err)
	defer cache.client.Close()
	defer func() {
		_, cleanupErr := cache.client.Del(context.Background(),
			socialFriendsCountKey(user1),
			socialFriendsCountKey(user2),
			socialBlockedCountKey(user1),
			socialBlockedCountKey(user2),
			socialCountLockKey(user1, "friends"),
			socialCountLockKey(user1, "blocked"),
			socialCountLockKey(user2, "friends"),
			socialCountLockKey(user2, "blocked"),
			socialCountVersionKey(socialFriendsCountKey(user1)),
			socialCountVersionKey(socialFriendsCountKey(user2)),
			socialCountVersionKey(socialBlockedCountKey(user1)),
			socialCountVersionKey(socialBlockedCountKey(user2)),
		).Result()
		require.NoError(t, cleanupErr)
	}()
	repo := &postgresFriendshipRepository{pool: pool, cache: cache}

	const parallelRequests = 6
	results := make(chan int, parallelRequests)
	errors := make(chan error, parallelRequests)
	for i := 0; i < parallelRequests; i++ {
		go func() {
			count, err := repo.GetAcceptedFriendsCount(context.Background(), user1)
			results <- count
			errors <- err
		}()
	}
	select {
	case <-tracer.firstFriendQuery:
	case <-time.After(5 * time.Second):
		t.Fatal("first PostgreSQL COUNT did not start")
	}
	time.Sleep(50 * time.Millisecond)
	close(tracer.releaseFirstFriendQuery)
	for i := 0; i < parallelRequests; i++ {
		require.NoError(t, <-errors)
		require.Zero(t, <-results)
	}
	friendQueries, _ := tracer.queryCounts()
	require.Equal(t, 1, friendQueries, "concurrent cache misses should share one PostgreSQL COUNT")

	count, err := repo.GetAcceptedFriendsCount(context.Background(), user1)
	require.NoError(t, err)
	require.Zero(t, count)
	friendQueries, _ = tracer.queryCounts()
	require.Equal(t, 1, friendQueries, "the next request should be served from Redis")
	ttl, err := cache.client.TTL(context.Background(), socialFriendsCountKey(user1)).Result()
	require.NoError(t, err)
	require.GreaterOrEqual(t, ttl, 9*time.Minute)
	require.LessOrEqual(t, ttl, socialCountTTL)

	for _, userID := range []string{user1, user2} {
		require.NoError(t, cache.setCount(context.Background(), socialFriendsCountKey(userID), 0, socialCountTTL))
		require.NoError(t, cache.setCount(context.Background(), socialBlockedCountKey(userID), 0, socialCountTTL))
	}
	request, err := repo.CreatePending(context.Background(), user1, user2)
	require.NoError(t, err)
	accepted, err := repo.Accept(context.Background(), string(request.ID), user2)
	require.NoError(t, err)
	require.Equal(t, models.FriendshipAccepted, accepted.Status)
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT estado FROM amistad WHERE id_amistad = $1`, string(request.ID)).Scan(&accepted.Status))
	require.Equal(t, models.FriendshipAccepted, accepted.Status)
	for _, userID := range []string{user1, user2} {
		require.False(t, socialCacheHasKey(t, cache, socialFriendsCountKey(userID)))
		require.False(t, socialCacheHasKey(t, cache, socialBlockedCountKey(userID)))
		count, err := repo.GetAcceptedFriendsCount(context.Background(), userID)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	}

	require.NoError(t, cache.invalidate(context.Background(), socialFriendsCountKey(user1)))
	type staleReadResult struct {
		count int
		err   error
	}
	staleCountCaptured := make(chan int, 1)
	releaseStaleRead := make(chan struct{})
	staleReadResultCh := make(chan staleReadResult, 1)
	go func() {
		count, err := repo.getSocialCount(
			context.Background(),
			socialFriendsCountKey(user1),
			socialCountLockKey(user1, "friends"),
			"friends",
			func(ctx context.Context) (int, error) {
				count, err := repo.queryAcceptedFriendsCount(ctx, user1)
				if err != nil {
					return 0, err
				}
				staleCountCaptured <- count
				<-releaseStaleRead
				return count, nil
			},
		)
		staleReadResultCh <- staleReadResult{count: count, err: err}
	}()
	select {
	case count := <-staleCountCaptured:
		require.Equal(t, 1, count)
	case <-time.After(5 * time.Second):
		t.Fatal("stale PostgreSQL count was not captured")
	}
	require.NoError(t, repo.RemoveFriend(context.Background(), user1, user2))
	require.False(t, socialCacheHasKey(t, cache, socialFriendsCountKey(user1)))
	close(releaseStaleRead)
	select {
	case result := <-staleReadResultCh:
		require.NoError(t, result.err)
		require.Zero(t, result.count, "the reader must retry after its stale publication is fenced")
	case <-time.After(5 * time.Second):
		t.Fatal("stale reader did not finish after invalidation")
	}
	for _, userID := range []string{user1, user2} {
		count, err := repo.GetAcceptedFriendsCount(context.Background(), userID)
		require.NoError(t, err)
		require.Zero(t, count)
	}

	for _, userID := range []string{user1, user2} {
		require.NoError(t, cache.setCount(context.Background(), socialFriendsCountKey(userID), 0, socialCountTTL))
		require.NoError(t, cache.setCount(context.Background(), socialBlockedCountKey(userID), 0, socialCountTTL))
	}
	require.NoError(t, repo.BlockUser(context.Background(), user1, user2))
	for _, userID := range []string{user1, user2} {
		require.False(t, socialCacheHasKey(t, cache, socialFriendsCountKey(userID)))
		require.False(t, socialCacheHasKey(t, cache, socialBlockedCountKey(userID)))
	}
	blockedCount, err := repo.GetBlockedUsersCount(context.Background(), user1)
	require.NoError(t, err)
	require.Equal(t, 1, blockedCount)

	require.NoError(t, repo.UnblockUser(context.Background(), user1, user2))
	require.False(t, socialCacheHasKey(t, cache, socialBlockedCountKey(user1)))
	blockedCount, err = repo.GetBlockedUsersCount(context.Background(), user1)
	require.NoError(t, err)
	require.Zero(t, blockedCount)

	brokenCache, err := newSocialCountRedisCache("redis://127.0.0.1:1")
	require.NoError(t, err)
	defer brokenCache.client.Close()
	brokenRepo := &postgresFriendshipRepository{pool: pool, cache: brokenCache}
	previousLogOutput := log.Writer()
	var warnings strings.Builder
	log.SetOutput(&warnings)
	count, err = brokenRepo.GetAcceptedFriendsCount(context.Background(), user1)
	log.SetOutput(previousLogOutput)
	require.NoError(t, err)
	require.Zero(t, count)
	require.Contains(t, warnings.String(), "WARN social-count-cache")
}

func socialCacheHasKey(t *testing.T, cache *socialCountRedisCache, key string) bool {
	t.Helper()
	exists, err := cache.client.Exists(context.Background(), key).Result()
	require.NoError(t, err)
	return exists > 0
}
