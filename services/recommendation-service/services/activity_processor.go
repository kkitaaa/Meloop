package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/meloop/recommendation-service/repositories"
)

type RecommendationInvalidator interface {
	MarkRecommendationsStale(ctx context.Context, userID string) error
}

type ActivityProcessor struct {
	activities repositories.RecentActivityRepository
	cache      RecommendationInvalidator
}

func NewActivityProcessor(activities repositories.RecentActivityRepository, cache RecommendationInvalidator) *ActivityProcessor {
	return &ActivityProcessor{activities: activities, cache: cache}
}

type InvalidActivityError struct {
	reason string
}

func (err *InvalidActivityError) Error() string {
	return err.reason
}

func (err *InvalidActivityError) Permanent() bool {
	return true
}

type activityPayload map[string]json.RawMessage

func (processor *ActivityProcessor) ProcessActivity(ctx context.Context, routingKey, messageID string, body []byte) error {
	interactionType, ok := map[string]string{
		"like.created":    "like",
		"comment.created": "comment",
		"post.created":    "post_interaction",
	}[routingKey]
	if !ok {
		return &InvalidActivityError{reason: fmt.Sprintf("unsupported activity event %q", routingKey)}
	}
	if processor.activities == nil || processor.cache == nil {
		return errors.New("recommendation activity dependencies are not initialized")
	}

	var payload activityPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return &InvalidActivityError{reason: fmt.Sprintf("decode activity event: %v", err)}
	}
	userID := firstIdentifier(payload, "userId", "user_id", "actorId", "actor_id", "authorId", "author_id", "creatorId", "creator_id", "id_usuario")
	postID := firstIdentifier(payload, "postId", "post_id", "publicationId", "publication_id", "id_publicacion", "target_id")
	if postID == "" && routingKey == "post.created" {
		postID = firstIdentifier(payload, "id")
	}
	if !validDatabaseIdentifier(userID) || !validDatabaseIdentifier(postID) {
		return &InvalidActivityError{reason: "activity event must contain valid user and post identifiers"}
	}

	eventID := strings.TrimSpace(messageID)
	if eventID == "" {
		eventID = firstIdentifier(payload, "eventId", "event_id", "likeId", "like_id", "commentId", "comment_id", "id")
	}
	if eventID == "" {
		digest := sha256.Sum256(append(append([]byte(routingKey), 0), body...))
		eventID = hex.EncodeToString(digest[:])
	}
	if len(eventID) > 64 || strings.ContainsRune(eventID, '\x00') {
		digest := sha256.Sum256([]byte(eventID))
		eventID = hex.EncodeToString(digest[:])
	}

	if err := processor.activities.RecordRecentActivity(ctx, repositories.RecentActivity{
		ID: eventID, UserID: userID, PostID: postID, Type: interactionType,
	}); err != nil {
		return fmt.Errorf("record recent user activity: %w", err)
	}
	if err := processor.cache.MarkRecommendationsStale(ctx, userID); err != nil {
		return fmt.Errorf("mark user recommendations stale: %w", err)
	}
	return nil
}

func validDatabaseIdentifier(value string) bool {
	return value != "" && len(value) <= 64 && !strings.ContainsRune(value, '\x00')
}

func firstIdentifier(payload activityPayload, keys ...string) string {
	for _, key := range keys {
		raw, exists := payload[key]
		if !exists || len(raw) == 0 {
			continue
		}
		var value string
		if err := json.Unmarshal(raw, &value); err == nil {
			if value = strings.TrimSpace(value); value != "" {
				return value
			}
			continue
		}
		var number json.Number
		if err := json.Unmarshal(raw, &number); err == nil {
			if value = number.String(); value != "" {
				return value
			}
		}
	}
	return ""
}

func (service *RecommendationService) MarkRecommendationsStale(ctx context.Context, userID string) error {
	if parsedUserID, err := strconv.Atoi(userID); err == nil {
		service.cache.InvalidateUser(parsedUserID)
	}
	if cacheStore, ok := service.store.(interface {
		InvalidateCachedForUser(context.Context, string) error
	}); ok {
		return cacheStore.InvalidateCachedForUser(ctx, userID)
	}
	if cacheStore, ok := service.store.(RecommendationCacheStore); ok {
		if parsedUserID, err := strconv.Atoi(userID); err == nil {
			return cacheStore.InvalidateCached(ctx, parsedUserID)
		}
	}
	return nil
}
