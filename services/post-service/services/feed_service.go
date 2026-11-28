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
	"strings"
	"sync"
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
	ErrInvalidFeedIdentity = errors.New("la identidad del feed no es válida o no coincide con la sesión")
)

type FeedIdentityVerifier interface {
	AuthenticatedUserID(ctx context.Context, authorization string) (string, error)
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

type localFeedSnapshot struct {
	snapshot models.FeedSnapshot
	expires  time.Time
}

type FeedService struct {
	repo            repositories.FeedRepository
	snapshots       repositories.FeedSnapshotStore
	blockedUsers    BlockedUsersProvider
	recommendations FeedRecommendationProvider
	localMu         sync.Mutex
	localSnapshots  map[string]localFeedSnapshot
}

func NewFeedService(
	repo repositories.FeedRepository,
	snapshots repositories.FeedSnapshotStore,
	blockedUsers BlockedUsersProvider,
	recommendations FeedRecommendationProvider,
) *FeedService {
	return &FeedService{
		repo:            repo,
		snapshots:       snapshots,
		blockedUsers:    blockedUsers,
		recommendations: recommendations,
		localSnapshots:  make(map[string]localFeedSnapshot),
	}
}

func (s *FeedService) GetFeed(ctx context.Context, userID string, limit int, cursor string) (FeedPage, error) {
	if limit < 1 || limit > maxFeedPageSize {
		return FeedPage{}, ErrInvalidFeedLimit
	}
	if userID == "" {
		return FeedPage{}, errors.New("el usuario es obligatorio")
	}
	if s.blockedUsers == nil {
		return FeedPage{}, errors.New("blocked users provider is not configured")
	}
	blockedAuthorIDs, err := s.blockedUsers.BlockedWith(ctx, userID)
	if err != nil {
		return FeedPage{}, fmt.Errorf("get blocked users: %w", err)
	}

	var snapshot models.FeedSnapshot
	offset := 0
	if cursor == "" {
		snapshot, err = s.createSnapshot(ctx, userID, blockedAuthorIDs)
		if err != nil {
			return FeedPage{}, err
		}
	} else {
		decoded, err := decodeFeedCursor(cursor)
		if err != nil {
			return FeedPage{}, err
		}
		snapshot, err = s.getSnapshot(ctx, userID)
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

	remainingIDs := snapshot.PostIDs[offset:]
	if len(remainingIDs) > maxFeedCandidates {
		return FeedPage{}, ErrInvalidFeedCursor
	}
	pageIDs := remainingIDs
	posts, err := s.repo.GetPostsByIDs(ctx, userID, pageIDs, blockedAuthorIDs)
	if err != nil {
		return FeedPage{}, err
	}

	posts, nextOffset := selectFeedPage(snapshot.PostIDs, offset, limit, posts)

	page := FeedPage{Posts: posts}
	if nextOffset < len(snapshot.PostIDs) {
		page.NextCursor, err = encodeFeedCursor(feedCursor{SnapshotID: snapshot.ID, Offset: nextOffset})
		if err != nil {
			return FeedPage{}, fmt.Errorf("codificar cursor del feed: %w", err)
		}
	}
	return page, nil
}

func (s *FeedService) createSnapshot(ctx context.Context, userID string, blockedAuthorIDs []string) (models.FeedSnapshot, error) {
	var recommendations []FeedRecommendation
	recommendationScores := make(map[string]float64)
	recommendedIDs := []string{}
	if s.recommendations != nil {
		items, err := s.recommendations.GetPostRecommendations(ctx, userID, maxFeedCandidates)
		if err != nil {
			slog.Warn("feed_recommendations_unavailable", "user_id", userID, "error", err)
		} else {
			if len(items) > maxFeedCandidates {
				items = items[:maxFeedCandidates]
			}
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

	candidates, err := s.repo.ListCandidates(ctx, userID, recommendedIDs, blockedAuthorIDs, maxFeedCandidates)
	if err != nil {
		return models.FeedSnapshot{}, err
	}
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
	key := feedSnapshotKey(userID)
	s.rememberSnapshot(key, snapshot)
	if s.snapshots != nil {
		if err := s.snapshots.Set(ctx, key, snapshot, feedSnapshotTTL); err != nil {
			slog.Warn("feed_snapshot_redis_write_failed", "user_id", userID, "error", err)
		}
	}
	return snapshot, nil
}

func (s *FeedService) getSnapshot(ctx context.Context, userID string) (models.FeedSnapshot, error) {
	key := feedSnapshotKey(userID)
	if s.snapshots != nil {
		snapshot, err := s.snapshots.Get(ctx, key)
		if err == nil {
			s.rememberSnapshot(key, snapshot)
			return snapshot, nil
		}
		if !errors.Is(err, redis.Nil) {
			slog.Warn("feed_snapshot_redis_read_failed", "user_id", userID, "error", err)
		}
		if localSnapshot, found := s.localSnapshot(key); found {
			return localSnapshot, nil
		}
		if errors.Is(err, redis.Nil) {
			return models.FeedSnapshot{}, ErrFeedSnapshotExpired
		}
		return models.FeedSnapshot{}, fmt.Errorf("%w: no local snapshot is available", ErrFeedSnapshotExpired)
	}
	if snapshot, found := s.localSnapshot(key); found {
		return snapshot, nil
	}
	return models.FeedSnapshot{}, ErrFeedSnapshotExpired
}

func (s *FeedService) rememberSnapshot(key string, snapshot models.FeedSnapshot) {
	s.localMu.Lock()
	defer s.localMu.Unlock()
	now := time.Now()
	if cached, exists := s.localSnapshots[key]; exists &&
		cached.snapshot.ID == snapshot.ID && now.Before(cached.expires) {
		return
	}
	for existingKey, cached := range s.localSnapshots {
		if !now.Before(cached.expires) {
			delete(s.localSnapshots, existingKey)
		}
	}
	if len(s.localSnapshots) >= 5000 {
		for existingKey := range s.localSnapshots {
			delete(s.localSnapshots, existingKey)
			break
		}
	}
	s.localSnapshots[key] = localFeedSnapshot{snapshot: snapshot, expires: now.Add(feedSnapshotTTL)}
}

func (s *FeedService) localSnapshot(key string) (models.FeedSnapshot, bool) {
	s.localMu.Lock()
	defer s.localMu.Unlock()
	cached, exists := s.localSnapshots[key]
	if !exists {
		return models.FeedSnapshot{}, false
	}
	if time.Now().After(cached.expires) {
		delete(s.localSnapshots, key)
		return models.FeedSnapshot{}, false
	}
	return cached.snapshot, true
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

func selectFeedPage(
	snapshotIDs []string,
	offset, limit int,
	posts []repositories.Post,
) ([]repositories.Post, int) {
	byID := make(map[string]repositories.Post, len(posts))
	for _, post := range posts {
		byID[post.ID] = post
	}
	ordered := make([]repositories.Post, 0, limit)
	nextOffset := offset
	for index := offset; index < len(snapshotIDs); index++ {
		nextOffset = index + 1
		id := snapshotIDs[index]
		if post, ok := byID[id]; ok {
			ordered = append(ordered, post)
			if len(ordered) == limit {
				break
			}
		}
	}
	return ordered, nextOffset
}

func feedSnapshotKey(userID string) string {
	return "feed:" + userID
}

func ValidateFeedUserID(userID string) bool {
	if userID == "" || len(userID) > 64 || strings.TrimSpace(userID) != userID {
		return false
	}
	for _, char := range userID {
		if char < 0x20 || char == 0x7f {
			return false
		}
	}
	return true
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
