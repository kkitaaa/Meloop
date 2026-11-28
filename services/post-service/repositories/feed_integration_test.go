package repositories

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/meloop/post-service/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgresFeedQueriesMatchSupabaseMigrations(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run feed integration tests against migrated Supabase PostgreSQL")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test PostgreSQL connection: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get test PostgreSQL connection: %v", err)
	}
	defer sqlDB.Close()

	repository := NewPostgresFeedRepository(db, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	userID := uuid.NewString()

	if _, err := repository.ListCandidates(ctx, userID, []string{uuid.NewString()}, maxFeedCandidates); err != nil {
		t.Fatalf("run feed candidate query against migrated schema: %v", err)
	}
	if _, err := repository.GetPostsByIDs(ctx, userID, []string{uuid.NewString()}); err != nil {
		t.Fatalf("run feed detail query against migrated schema: %v", err)
	}
	if _, err := repository.ExcludeBlockedAuthors(ctx, userID, []string{uuid.NewString()}); err != nil {
		t.Fatalf("run PostgreSQL block fallback query against migrated schema: %v", err)
	}
}

func TestRedisFeedSnapshotStoreReportsUnavailableRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:0",
		DialTimeout:  20 * time.Millisecond,
		ReadTimeout:  20 * time.Millisecond,
		WriteTimeout: 20 * time.Millisecond,
	})
	defer client.Close()

	store := NewRedisFeedSnapshotStore(client, 20*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := store.Get(ctx, "feed:integration-test"); err == nil {
		t.Fatal("expected Redis read to fail when Redis is unavailable")
	}
	if err := store.Set(ctx, "feed:integration-test", models.FeedSnapshot{ID: uuid.NewString()}, time.Minute); err == nil {
		t.Fatal("expected Redis write to fail when Redis is unavailable")
	}
}
