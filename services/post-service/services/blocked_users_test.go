package services

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type blockedUsersCacheFake struct {
	values map[string]string
	ttl    time.Duration
	getErr error
	setErr error
}

func (c *blockedUsersCacheFake) Get(_ context.Context, key string) (string, error) {
	if c.getErr != nil {
		return "", c.getErr
	}
	value, ok := c.values[key]
	if !ok {
		return "", redis.Nil
	}
	return value, nil
}

func (c *blockedUsersCacheFake) Set(_ context.Context, key, value string, ttl time.Duration) error {
	if c.setErr != nil {
		return c.setErr
	}
	if c.values == nil {
		c.values = make(map[string]string)
	}
	c.values[key] = value
	c.ttl = ttl
	return nil
}

type blockedUsersSourceFake struct {
	ids   []string
	calls int
	err   error
}

func (s *blockedUsersSourceFake) ListUserIDs(context.Context) ([]string, error) {
	s.calls++
	return s.ids, s.err
}

type blockedCandidatesFake struct {
	blocked []string
	calls   int
	userID  string
	input   []string
	err     error
}

func (s *blockedCandidatesFake) GetBlockedCandidateIDs(_ context.Context, userID string, candidates []string) ([]string, error) {
	s.calls++
	s.userID = userID
	s.input = append([]string(nil), candidates...)
	return s.blocked, s.err
}

func TestBlockedWithUsesSocialServiceAndCachesForOneMinute(t *testing.T) {
	cache := &blockedUsersCacheFake{}
	users := &blockedUsersSourceFake{ids: []string{"user-a", "user-b", "viewer"}}
	social := &blockedCandidatesFake{blocked: []string{"user-a", "user-b"}}
	provider := newRedisBlockedUsers(cache, users, social)

	for call := 0; call < 2; call++ {
		blocked, err := provider.BlockedWith(context.Background(), "viewer")
		if err != nil {
			t.Fatalf("BlockedWith call %d: %v", call+1, err)
		}
		if !reflect.DeepEqual(blocked, []string{"user-a", "user-b"}) {
			t.Fatalf("blocked users = %v, want [user-a user-b]", blocked)
		}
	}
	if social.calls != 1 || users.calls != 1 {
		t.Fatalf("social calls = %d and user-list calls = %d, want one each", social.calls, users.calls)
	}
	if social.userID != "viewer" || !reflect.DeepEqual(social.input, users.ids) {
		t.Fatalf("social request user=%q candidates=%v", social.userID, social.input)
	}
	if cache.ttl != 60*time.Second {
		t.Fatalf("cache TTL = %s, want 60s", cache.ttl)
	}
}

func TestBlockedWithPropagatesSocialErrors(t *testing.T) {
	provider := newRedisBlockedUsers(
		&blockedUsersCacheFake{},
		&blockedUsersSourceFake{ids: []string{"author"}},
		&blockedCandidatesFake{err: errors.New("social-service unavailable")},
	)
	if _, err := provider.BlockedWith(context.Background(), "viewer"); err == nil {
		t.Fatal("expected blocked-user lookup error")
	}
}

func TestBlockedWithFallsBackToSocialWhenRedisIsUnavailable(t *testing.T) {
	social := &blockedCandidatesFake{blocked: []string{"author"}}
	provider := newRedisBlockedUsers(
		&blockedUsersCacheFake{getErr: errors.New("redis unavailable"), setErr: errors.New("redis unavailable")},
		&blockedUsersSourceFake{ids: []string{"author", "viewer"}},
		social,
	)

	blocked, err := provider.BlockedWith(context.Background(), "viewer")
	if err != nil {
		t.Fatalf("BlockedWith should work without Redis: %v", err)
	}
	if !reflect.DeepEqual(blocked, []string{"author"}) {
		t.Fatalf("blocked users = %v, want [author]", blocked)
	}
	if social.calls != 1 {
		t.Fatalf("social calls = %d, want 1", social.calls)
	}
}

func TestBlockedWithReplacesMalformedCacheEntry(t *testing.T) {
	key := blockedUsersCacheKey("viewer")
	cache := &blockedUsersCacheFake{values: map[string]string{key: "invalid json"}}
	provider := newRedisBlockedUsers(cache, &blockedUsersSourceFake{ids: []string{"author"}}, &blockedCandidatesFake{blocked: []string{"author"}})

	blocked, err := provider.BlockedWith(context.Background(), "viewer")
	if err != nil {
		t.Fatalf("BlockedWith with malformed cache: %v", err)
	}
	if !reflect.DeepEqual(blocked, []string{"author"}) {
		t.Fatalf("blocked users = %v, want [author]", blocked)
	}
	if cache.values[key] == "invalid json" {
		t.Fatal("expected malformed cache entry to be replaced")
	}
}
