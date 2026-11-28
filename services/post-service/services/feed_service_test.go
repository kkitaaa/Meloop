package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/redis/go-redis/v9"
)

type feedRepositoryFake struct {
	candidates []repositories.FeedCandidate
	posts      map[string]repositories.Post
	blocked    map[string]bool
	blockErr   error
}

func (repo *feedRepositoryFake) ListCandidates(_ context.Context, _ string, _ []string, blockedAuthorIDs []string, _ int) ([]repositories.FeedCandidate, error) {
	blocked := stringSet(blockedAuthorIDs)
	candidates := make([]repositories.FeedCandidate, 0, len(repo.candidates))
	for _, candidate := range repo.candidates {
		if _, excluded := blocked[candidate.AuthorID]; !excluded {
			candidates = append(candidates, candidate)
		}
	}
	return candidates, nil
}

func (repo *feedRepositoryFake) GetPostsByIDs(_ context.Context, _ string, postIDs, blockedAuthorIDs []string) ([]repositories.Post, error) {
	posts := make([]repositories.Post, 0, len(postIDs))
	blocked := stringSet(blockedAuthorIDs)
	for _, postID := range postIDs {
		if post, ok := repo.posts[postID]; ok {
			if _, excluded := blocked[post.AuthorID]; !excluded {
				posts = append(posts, post)
			}
		}
	}
	return posts, nil
}

type feedSnapshotStoreFake struct {
	snapshot models.FeedSnapshot
	getErr   error
	setErr   error
}

func (store *feedSnapshotStoreFake) Get(context.Context, string) (models.FeedSnapshot, error) {
	if store.getErr != nil {
		return models.FeedSnapshot{}, store.getErr
	}
	if store.snapshot.ID == "" {
		return models.FeedSnapshot{}, redis.Nil
	}
	return store.snapshot, nil
}

func (store *feedSnapshotStoreFake) Set(_ context.Context, _ string, snapshot models.FeedSnapshot, ttl time.Duration) error {
	if ttl != feedSnapshotTTL {
		return errors.New("unexpected feed snapshot TTL")
	}
	if store.setErr != nil {
		return store.setErr
	}
	store.snapshot = snapshot
	return nil
}

type feedBlockFilterFake struct {
	blocked map[string]bool
	err     error
}

func (filter feedBlockFilterFake) BlockedWith(_ context.Context, _ string) ([]string, error) {
	if filter.err != nil {
		return nil, filter.err
	}
	blocked := make([]string, 0, len(filter.blocked))
	for authorID, isBlocked := range filter.blocked {
		if isBlocked {
			blocked = append(blocked, authorID)
		}
	}
	return blocked, nil
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
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

func TestGetFeedFailsClosedWhenBlockedUsersUnavailable(t *testing.T) {
	service := NewFeedService(
		&feedRepositoryFake{},
		&feedSnapshotStoreFake{},
		feedBlockFilterFake{err: errors.New("social-service unavailable")},
		nil,
	)
	if _, err := service.GetFeed(context.Background(), "viewer", 20, ""); err == nil {
		t.Fatal("expected an error when blocked users cannot be loaded")
	}
}

func TestGetFeedContinuesWhenRedisIsUnavailable(t *testing.T) {
	now := time.Now()
	repo := &feedRepositoryFake{
		candidates: []repositories.FeedCandidate{
			{Post: repositories.Post{ID: "post-a", AuthorID: "author-a", CreatedAt: now}},
			{Post: repositories.Post{ID: "post-b", AuthorID: "author-b", CreatedAt: now.Add(-time.Minute)}},
		},
		posts: map[string]repositories.Post{
			"post-a": {ID: "post-a", AuthorID: "author-a"},
			"post-b": {ID: "post-b", AuthorID: "author-b"},
		},
	}
	store := &feedSnapshotStoreFake{
		getErr: errors.New("redis unavailable"),
		setErr: errors.New("redis unavailable"),
	}
	recommendationServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "unavailable", http.StatusServiceUnavailable)
	}))
	defer recommendationServer.Close()
	service := NewFeedService(
		repo,
		store,
		feedBlockFilterFake{},
		NewHTTPFeedRecommendationProvider(recommendationServer.URL, time.Second),
	)

	firstPage, err := service.GetFeed(context.Background(), "viewer", 1, "")
	if err != nil {
		t.Fatalf("get first page while Redis is unavailable: %v", err)
	}
	if len(firstPage.Posts) != 1 || firstPage.NextCursor == "" {
		t.Fatalf("first page = %+v, want one post and a cursor", firstPage)
	}

	secondPage, err := service.GetFeed(context.Background(), "viewer", 1, firstPage.NextCursor)
	if err != nil {
		t.Fatalf("get next page from in-memory snapshot: %v", err)
	}
	if len(secondPage.Posts) != 1 || secondPage.Posts[0].ID == firstPage.Posts[0].ID {
		t.Fatalf("second page = %+v, want the next distinct post", secondPage)
	}
}

func TestGetFeedSkipsPostsDeletedAfterSnapshot(t *testing.T) {
	now := time.Now()
	repo := &feedRepositoryFake{
		candidates: []repositories.FeedCandidate{
			{Post: repositories.Post{ID: "first", AuthorID: "author-first", CreatedAt: now}},
			{Post: repositories.Post{ID: "deleted", AuthorID: "author-deleted", CreatedAt: now.Add(-time.Minute)}},
			{Post: repositories.Post{ID: "later", AuthorID: "author-later", CreatedAt: now.Add(-2 * time.Minute)}},
		},
		posts: map[string]repositories.Post{
			"first":   {ID: "first", AuthorID: "author-first"},
			"deleted": {ID: "deleted", AuthorID: "author-deleted"},
			"later":   {ID: "later", AuthorID: "author-later"},
		},
	}
	store := &feedSnapshotStoreFake{}
	service := NewFeedService(repo, store, feedBlockFilterFake{}, nil)

	firstPage, err := service.GetFeed(context.Background(), "viewer", 1, "")
	if err != nil {
		t.Fatalf("get first page: %v", err)
	}
	delete(repo.posts, "deleted")

	secondPage, err := service.GetFeed(context.Background(), "viewer", 1, firstPage.NextCursor)
	if err != nil {
		t.Fatalf("get page after a post was deleted: %v", err)
	}
	if len(secondPage.Posts) != 1 || secondPage.Posts[0].ID != "later" {
		t.Fatalf("second page = %+v, want later post after skipping deleted post", secondPage.Posts)
	}
	if secondPage.NextCursor != "" {
		t.Fatalf("next cursor = %q, want empty", secondPage.NextCursor)
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
