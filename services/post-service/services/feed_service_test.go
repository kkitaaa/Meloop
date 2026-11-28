package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/redis/go-redis/v9"
)

type feedRepositoryFake struct {
	candidates []repositories.FeedCandidate
	posts      map[string]repositories.Post
}

func (repo *feedRepositoryFake) ListCandidates(context.Context, string, []string, int) ([]repositories.FeedCandidate, error) {
	return repo.candidates, nil
}

func (repo *feedRepositoryFake) GetPostsByIDs(_ context.Context, _ string, postIDs []string) ([]repositories.Post, error) {
	posts := make([]repositories.Post, 0, len(postIDs))
	for _, postID := range postIDs {
		if post, ok := repo.posts[postID]; ok {
			posts = append(posts, post)
		}
	}
	return posts, nil
}

type feedSnapshotStoreFake struct {
	snapshot models.FeedSnapshot
}

func (store *feedSnapshotStoreFake) Get(context.Context, string) (models.FeedSnapshot, error) {
	if store.snapshot.ID == "" {
		return models.FeedSnapshot{}, redis.Nil
	}
	return store.snapshot, nil
}

func (store *feedSnapshotStoreFake) Set(_ context.Context, _ string, snapshot models.FeedSnapshot, ttl time.Duration) error {
	if ttl != feedSnapshotTTL {
		return errors.New("unexpected feed snapshot TTL")
	}
	store.snapshot = snapshot
	return nil
}

type feedBlockFilterFake struct {
	blocked map[string]bool
}

func (filter feedBlockFilterFake) ExcludeBlocked(_ context.Context, _ string, authorIDs []string) ([]string, error) {
	allowed := make([]string, 0, len(authorIDs))
	for _, authorID := range authorIDs {
		if !filter.blocked[authorID] {
			allowed = append(allowed, authorID)
		}
	}
	return allowed, nil
}

type feedRecommendationFake []FeedRecommendation

func (recommendations feedRecommendationFake) GetPostRecommendations(context.Context, string, int) ([]FeedRecommendation, error) {
	return recommendations, nil
}

func TestGetFeedKeepsSnapshotOrderAndFiltersBlockedAuthors(t *testing.T) {
	now := time.Now()
	repo := &feedRepositoryFake{
		candidates: []repositories.FeedCandidate{
			{Post: repositories.Post{ID: "post-a", AuthorID: "author-a", CreatedAt: now.Add(-time.Hour)}, CommentsCount: 1},
			{Post: repositories.Post{ID: "post-a", AuthorID: "author-a", CreatedAt: now.Add(-time.Hour)}, CommentsCount: 1},
			{Post: repositories.Post{ID: "post-b", AuthorID: "author-b", CreatedAt: now.Add(-2 * time.Hour)}},
			{Post: repositories.Post{ID: "post-blocked", AuthorID: "author-blocked", CreatedAt: now}},
		},
		posts: map[string]repositories.Post{
			"post-a":       {ID: "post-a", AuthorID: "author-a"},
			"post-b":       {ID: "post-b", AuthorID: "author-b"},
			"post-blocked": {ID: "post-blocked", AuthorID: "author-blocked"},
		},
	}
	store := &feedSnapshotStoreFake{}
	service := NewFeedService(
		repo,
		store,
		feedBlockFilterFake{blocked: map[string]bool{"author-blocked": true}},
		feedRecommendationFake{{PostID: "post-b", Score: 2}},
	)

	firstPage, err := service.GetFeed(context.Background(), "viewer", 1, "")
	if err != nil {
		t.Fatalf("get first feed page: %v", err)
	}
	if len(firstPage.Posts) != 1 || firstPage.Posts[0].ID != "post-b" {
		t.Fatalf("first page = %#v, want recommended post-b", firstPage.Posts)
	}
	if firstPage.NextCursor == "" {
		t.Fatal("expected a cursor for the next page")
	}
	if got, want := store.snapshot.PostIDs, []string{"post-b", "post-a"}; !equalStrings(got, want) {
		t.Fatalf("snapshot IDs = %v, want %v", got, want)
	}

	repo.candidates = append(repo.candidates,
		repositories.FeedCandidate{Post: repositories.Post{ID: "new-post", AuthorID: "author-new", CreatedAt: now.Add(time.Minute)}})
	repo.posts["new-post"] = repositories.Post{ID: "new-post", AuthorID: "author-new"}

	secondPage, err := service.GetFeed(context.Background(), "viewer", 1, firstPage.NextCursor)
	if err != nil {
		t.Fatalf("get second feed page: %v", err)
	}
	if len(secondPage.Posts) != 1 || secondPage.Posts[0].ID != "post-a" {
		t.Fatalf("second page = %#v, want post-a from the original snapshot", secondPage.Posts)
	}
	if secondPage.NextCursor != "" {
		t.Fatalf("last page cursor = %q, want empty", secondPage.NextCursor)
	}
}

func TestGetFeedRejectsExpiredAndMalformedCursors(t *testing.T) {
	service := NewFeedService(
		&feedRepositoryFake{},
		&feedSnapshotStoreFake{},
		feedBlockFilterFake{},
		nil,
	)
	if _, err := service.GetFeed(context.Background(), "viewer", 20, "not-a-cursor"); !errors.Is(err, ErrInvalidFeedCursor) {
		t.Fatalf("malformed cursor error = %v, want ErrInvalidFeedCursor", err)
	}

	cursor, err := encodeFeedCursor(feedCursor{SnapshotID: "expired", Offset: 0})
	if err != nil {
		t.Fatalf("encode test cursor: %v", err)
	}
	if _, err := service.GetFeed(context.Background(), "viewer", 20, cursor); !errors.Is(err, ErrFeedSnapshotExpired) {
		t.Fatalf("expired cursor error = %v, want ErrFeedSnapshotExpired", err)
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
