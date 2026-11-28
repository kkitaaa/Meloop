package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/redis/go-redis/v9"
)

const (
	feedSnapshotTTL   = 2 * time.Minute
	maxFeedPageSize   = 50
	maxFeedCandidates = 200
)

var (
	ErrFeedSnapshotExpired = errors.New("el cursor del feed expiró; inicia una nueva consulta")
	ErrInvalidFeedCursor   = errors.New("el cursor del feed no es válido")
	ErrInvalidFeedLimit    = errors.New("limit debe ser un entero entre 1 y 50")
)

type FeedBlockFilter interface {
	ExcludeBlocked(ctx context.Context, userID string, authorIDs []string) ([]string, error)
}

type FeedRecommendation struct {
	PostID string
	Score  float64
}

type FeedRecommendationProvider interface {
	GetPostRecommendations(ctx context.Context, userID string, limit int) ([]FeedRecommendation, error)
}

type FeedPage struct {
	Posts      []repositories.Post
	NextCursor string
}

type feedCursor struct {
	SnapshotID string `json:"snapshot_id"`
	Offset     int    `json:"offset"`
}

type FeedService struct {
	repo            repositories.FeedRepository
	snapshots       repositories.FeedSnapshotStore
	blockFilter     FeedBlockFilter
	recommendations FeedRecommendationProvider
}

func NewFeedService(
	repo repositories.FeedRepository,
	snapshots repositories.FeedSnapshotStore,
	blockFilter FeedBlockFilter,
	recommendations FeedRecommendationProvider,
) *FeedService {
	return &FeedService{
		repo:            repo,
		snapshots:       snapshots,
		blockFilter:     blockFilter,
		recommendations: recommendations,
	}
}

func (s *FeedService) GetFeed(ctx context.Context, userID string, limit int, cursor string) (FeedPage, error) {
	if limit < 1 || limit > maxFeedPageSize {
		return FeedPage{}, ErrInvalidFeedLimit
	}
	if userID == "" {
		return FeedPage{}, errors.New("el usuario es obligatorio")
	}

	ctx, cancel := context.WithTimeout(ctx, 450*time.Millisecond)
	defer cancel()

	var snapshot models.FeedSnapshot
	offset := 0
	if cursor == "" {
		var err error
		snapshot, err = s.createSnapshot(ctx, userID)
		if err != nil {
			return FeedPage{}, err
		}
	} else {
		decoded, err := decodeFeedCursor(cursor)
		if err != nil {
			return FeedPage{}, err
		}
		snapshot, err = s.snapshots.Get(ctx, feedSnapshotKey(userID))
		if errors.Is(err, redis.Nil) {
			return FeedPage{}, ErrFeedSnapshotExpired
		}
		if err != nil {
			return FeedPage{}, err
		}
		if snapshot.ID != decoded.SnapshotID {
			return FeedPage{}, ErrFeedSnapshotExpired
		}
		offset = decoded.Offset
		if offset < 0 || offset > len(snapshot.PostIDs) {
			return FeedPage{}, ErrInvalidFeedCursor
		}
	}

	if offset == len(snapshot.PostIDs) {
		return FeedPage{Posts: []repositories.Post{}}, nil
	}

	end := offset + limit
	if end > len(snapshot.PostIDs) {
		end = len(snapshot.PostIDs)
	}
	pageIDs := snapshot.PostIDs[offset:end]
	posts, err := s.repo.GetPostsByIDs(ctx, userID, pageIDs)
	if err != nil {
		return FeedPage{}, err
	}

	allowedAuthors, err := s.allowedAuthors(ctx, userID, posts)
	if err != nil {
		return FeedPage{}, err
	}
	posts = filterAndOrderPosts(posts, pageIDs, allowedAuthors)

	page := FeedPage{Posts: posts}
	if end < len(snapshot.PostIDs) {
		page.NextCursor, err = encodeFeedCursor(feedCursor{SnapshotID: snapshot.ID, Offset: end})
		if err != nil {
			return FeedPage{}, fmt.Errorf("codificar cursor del feed: %w", err)
		}
	}
	return page, nil
}

func (s *FeedService) createSnapshot(ctx context.Context, userID string) (models.FeedSnapshot, error) {
	var recommendations []FeedRecommendation
	recommendationScores := make(map[string]float64)
	recommendedIDs := []string{}
	if s.recommendations != nil {
		items, err := s.recommendations.GetPostRecommendations(ctx, userID, maxFeedCandidates)
		if err != nil {
			slog.Warn("feed_recommendations_unavailable", "user_id", userID, "error", err)
		} else {
			recommendations = items
			for _, item := range recommendations {
				if item.PostID == "" {
					continue
				}
				if previousScore, exists := recommendationScores[item.PostID]; exists {
					if item.Score > previousScore {
						recommendationScores[item.PostID] = item.Score
					}
				} else {
					recommendedIDs = append(recommendedIDs, item.PostID)
					recommendationScores[item.PostID] = item.Score
				}
			}
		}
	}

	candidates, err := s.repo.ListCandidates(ctx, userID, recommendedIDs, maxFeedCandidates)
	if err != nil {
		return models.FeedSnapshot{}, err
	}
	allowedAuthors, err := s.allowedAuthors(ctx, userID, candidatePosts(candidates))
	if err != nil {
		return models.FeedSnapshot{}, err
	}

	candidates = filterCandidates(candidates, allowedAuthors)
	sort.SliceStable(candidates, func(i, j int) bool {
		left := feedScore(candidates[i], recommendationScores[candidates[i].ID])
		right := feedScore(candidates[j], recommendationScores[candidates[j].ID])
		if left != right {
			return left > right
		}
		if !candidates[i].CreatedAt.Equal(candidates[j].CreatedAt) {
			return candidates[i].CreatedAt.After(candidates[j].CreatedAt)
		}
		return candidates[i].ID < candidates[j].ID
	})
	if len(candidates) > maxFeedCandidates {
		candidates = candidates[:maxFeedCandidates]
	}

	postIDs := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate.ID == "" {
			continue
		}
		if _, exists := seen[candidate.ID]; exists {
			continue
		}
		seen[candidate.ID] = struct{}{}
		postIDs = append(postIDs, candidate.ID)
	}

	snapshot := models.FeedSnapshot{ID: uuid.NewString(), PostIDs: postIDs}
	if err := s.snapshots.Set(ctx, feedSnapshotKey(userID), snapshot, feedSnapshotTTL); err != nil {
		return models.FeedSnapshot{}, err
	}
	return snapshot, nil
}

func (s *FeedService) allowedAuthors(ctx context.Context, userID string, posts []repositories.Post) (map[string]struct{}, error) {
	authorIDs := make([]string, 0, len(posts))
	seen := make(map[string]struct{}, len(posts))
	for _, post := range posts {
		if post.AuthorID == "" {
			continue
		}
		if _, exists := seen[post.AuthorID]; exists {
			continue
		}
		seen[post.AuthorID] = struct{}{}
		authorIDs = append(authorIDs, post.AuthorID)
	}
	if len(authorIDs) == 0 {
		return map[string]struct{}{}, nil
	}
	allowed, err := s.blockFilter.ExcludeBlocked(ctx, userID, authorIDs)
	if err != nil {
		return nil, err
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, authorID := range allowed {
		allowedSet[authorID] = struct{}{}
	}
	return allowedSet, nil
}

func feedScore(candidate repositories.FeedCandidate, recommendationScore float64) float64 {
	ageHours := time.Since(candidate.CreatedAt).Hours()
	if ageHours < 0 {
		ageHours = 0
	}
	if ageHours > 7*24 {
		ageHours = 7 * 24
	}
	recency := 1 - ageHours/(7*24)
	engagement := math.Log1p(float64(candidate.LikesCount + 2*candidate.CommentsCount))
	return recency + engagement + recommendationScore
}

func candidatePosts(candidates []repositories.FeedCandidate) []repositories.Post {
	posts := make([]repositories.Post, len(candidates))
	for index := range candidates {
		posts[index] = candidates[index].Post
	}
	return posts
}

func filterCandidates(candidates []repositories.FeedCandidate, allowedAuthors map[string]struct{}) []repositories.FeedCandidate {
	filtered := make([]repositories.FeedCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if _, ok := allowedAuthors[candidate.AuthorID]; ok {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func filterAndOrderPosts(posts []repositories.Post, requestedIDs []string, allowedAuthors map[string]struct{}) []repositories.Post {
	byID := make(map[string]repositories.Post, len(posts))
	for _, post := range posts {
		if _, ok := allowedAuthors[post.AuthorID]; ok {
			byID[post.ID] = post
		}
	}
	ordered := make([]repositories.Post, 0, len(posts))
	for _, id := range requestedIDs {
		if post, ok := byID[id]; ok {
			ordered = append(ordered, post)
		}
	}
	return ordered
}

func feedSnapshotKey(userID string) string {
	return "feed:" + userID
}

func encodeFeedCursor(cursor feedCursor) (string, error) {
	encoded, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func decodeFeedCursor(value string) (feedCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return feedCursor{}, ErrInvalidFeedCursor
	}
	var cursor feedCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil || cursor.SnapshotID == "" || cursor.Offset < 0 {
		return feedCursor{}, ErrInvalidFeedCursor
	}
	return cursor, nil
}

func parseRecommendationUserID(userID string) (int, bool) {
	id, err := strconv.Atoi(userID)
	return id, err == nil && id > 0
}
