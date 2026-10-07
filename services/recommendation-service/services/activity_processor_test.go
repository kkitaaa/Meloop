package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/recommendation-service/repositories"
)

type recentActivityRepositoryFake struct {
	activities []repositories.RecentActivity
	err        error
}

func (repository *recentActivityRepositoryFake) RecordRecentActivity(_ context.Context, activity repositories.RecentActivity) error {
	if repository.err != nil {
		return repository.err
	}
	repository.activities = append(repository.activities, activity)
	return nil
}

type recommendationInvalidatorFake struct {
	userIDs []string
	err     error
}

func (invalidator *recommendationInvalidatorFake) MarkRecommendationsStale(_ context.Context, userID string) error {
	if invalidator.err != nil {
		return invalidator.err
	}
	invalidator.userIDs = append(invalidator.userIDs, userID)
	return nil
}

func TestActivityProcessorPersistsSupportedEventsAndInvalidatesRecommendations(t *testing.T) {
	tests := []struct {
		routingKey string
		body       string
		wantType   string
		wantPostID string
	}{
		{"like.created", `{"userId":"12","postId":"34","likeId":"like-1"}`, "like", "34"},
		{"comment.created", `{"user_id":12,"post_id":34,"comment_id":"comment-1"}`, "comment", "34"},
		{"post.created", `{"author_id":"12","id":"34"}`, "post_interaction", "34"},
	}
	for _, test := range tests {
		t.Run(test.routingKey, func(t *testing.T) {
			repository := &recentActivityRepositoryFake{}
			invalidator := &recommendationInvalidatorFake{}
			processor := NewActivityProcessor(repository, invalidator)

			if err := processor.ProcessActivity(context.Background(), test.routingKey, "", []byte(test.body)); err != nil {
				t.Fatalf("process event: %v", err)
			}
			if len(repository.activities) != 1 {
				t.Fatalf("expected one persisted activity, got %d", len(repository.activities))
			}
			activity := repository.activities[0]
			if activity.UserID != "12" || activity.PostID != test.wantPostID || activity.Type != test.wantType {
				t.Fatalf("unexpected persisted activity: %+v", activity)
			}
			if activity.ID == "" || len(activity.ID) > 64 {
				t.Fatalf("expected a non-empty database-safe event id, got %q", activity.ID)
			}
			if len(invalidator.userIDs) != 1 || invalidator.userIDs[0] != "12" {
				t.Fatalf("expected recommendations to be invalidated for user 12, got %v", invalidator.userIDs)
			}
		})
	}
}

func TestActivityProcessorUsesBrokerMessageIDAndInvalidatesAfterPersistence(t *testing.T) {
	repository := &recentActivityRepositoryFake{}
	invalidator := &recommendationInvalidatorFake{err: errors.New("Redis unavailable")}
	processor := NewActivityProcessor(repository, invalidator)

	err := processor.ProcessActivity(
		context.Background(),
		"like.created",
		"broker-event-1",
		[]byte(`{"userId":"12","postId":"34"}`),
	)
	if err == nil {
		t.Fatal("expected Redis invalidation failure to be returned for message retry")
	}
	if len(repository.activities) != 1 || repository.activities[0].ID != "broker-event-1" {
		t.Fatalf("expected persisted idempotency key from RabbitMQ, got %+v", repository.activities)
	}
}

func TestActivityProcessorRejectsMalformedEventsWithoutPersistence(t *testing.T) {
	repository := &recentActivityRepositoryFake{}
	invalidator := &recommendationInvalidatorFake{}
	processor := NewActivityProcessor(repository, invalidator)

	err := processor.ProcessActivity(context.Background(), "like.created", "", []byte(`{"userId":"12"}`))
	var invalid *InvalidActivityError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected invalid activity error, got %v", err)
	}
	if len(repository.activities) != 0 || len(invalidator.userIDs) != 0 {
		t.Fatalf("invalid event caused side effects: activities=%+v invalidations=%v", repository.activities, invalidator.userIDs)
	}
}
