package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/notification-service/models"
)

type fakeRepository struct {
	postOwner        string
	interactionOwner string
	notifications    []models.Notification
}

func (r *fakeRepository) ResolvePostOwner(context.Context, string) (string, error) {
	return r.postOwner, nil
}

func (r *fakeRepository) ResolveInteractionOwner(context.Context, string) (string, error) {
	return r.interactionOwner, nil
}

func (r *fakeRepository) Create(_ context.Context, notification models.Notification) error {
	r.notifications = append(r.notifications, notification)
	return nil
}

func TestProcessPostLikedResolvesPostOwner(t *testing.T) {
	repository := &fakeRepository{postOwner: "post-owner"}
	processor := NewProcessor(repository)
	body := []byte(`{"event":"PostLiked","userId":"liker","postId":"post-1"}`)

	if err := processor.Process(context.Background(), "post.liked", "event-1", body); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(repository.notifications) != 1 {
		t.Fatalf("notification count = %d, want 1", len(repository.notifications))
	}
	notification := repository.notifications[0]
	if notification.Recipient != "post-owner" || notification.Type != "post.liked" || notification.EventID != "event-1" {
		t.Fatalf("unexpected notification: %+v", notification)
	}
	if notification.Actor == nil || *notification.Actor != "liker" {
		t.Fatalf("actor = %v, want liker", notification.Actor)
	}
}

func TestProcessFriendAcceptedNotifiesOriginalSender(t *testing.T) {
	repository := &fakeRepository{}
	processor := NewProcessor(repository)
	body := []byte(`{"event":"friend.accepted","sender_id":"requester","receiver_id":"acceptor"}`)

	if err := processor.Process(context.Background(), "friend.accepted", "", body); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	notification := repository.notifications[0]
	if notification.Recipient != "requester" || notification.Actor == nil || *notification.Actor != "acceptor" {
		t.Fatalf("unexpected notification: %+v", notification)
	}
	if notification.EventID == "" {
		t.Fatal("expected generated idempotency key")
	}
}

func TestProcessCommentReplyResolvesParentCommentOwner(t *testing.T) {
	repository := &fakeRepository{interactionOwner: "comment-owner"}
	processor := NewProcessor(repository)
	body := []byte(`{"user_id":"replier","reply_to_id":"comment-1"}`)

	if err := processor.Process(context.Background(), "comment.replied", "reply-event", body); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	notification := repository.notifications[0]
	if notification.Recipient != "comment-owner" || notification.Type != eventCommentReplied {
		t.Fatalf("unexpected notification: %+v", notification)
	}
}

func TestProcessUsesExplicitRecipientAndPreservesSystemEvent(t *testing.T) {
	repository := &fakeRepository{}
	processor := NewProcessor(repository)
	body := []byte(`{"user_id":"levelled-user","new_level":4}`)

	if err := processor.Process(context.Background(), "user.level_up", "event-level", body); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if got := repository.notifications[0].Recipient; got != "levelled-user" {
		t.Fatalf("recipient = %q, want levelled-user", got)
	}
	if repository.notifications[0].Actor != nil {
		t.Fatalf("system notification actor = %v, want nil", repository.notifications[0].Actor)
	}
}

func TestProcessRejectsMalformedAndUnsupportedEvents(t *testing.T) {
	processor := NewProcessor(&fakeRepository{})
	for name, body := range map[string][]byte{
		"malformed":   []byte(`{`),
		"unsupported": []byte(`{"event":"post.deleted"}`),
	} {
		t.Run(name, func(t *testing.T) {
			err := processor.Process(context.Background(), "", "", body)
			if err == nil || !IsPermanent(err) {
				t.Fatalf("Process() error = %v, want permanent error", err)
			}
		})
	}
}

func TestProcessPropagatesRepositoryFailureAsRetryable(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repository := &errorRepository{fakeRepository: fakeRepository{postOwner: "owner"}, err: wantErr}
	processor := NewProcessor(repository)
	err := processor.Process(context.Background(), "post.liked", "event-1", []byte(`{"userId":"liker","postId":"post-1"}`))
	if !errors.Is(err, wantErr) || IsPermanent(err) {
		t.Fatalf("Process() error = %v, want retryable repository error", err)
	}
}

type errorRepository struct {
	fakeRepository
	err error
}

func (r *errorRepository) ResolvePostOwner(context.Context, string) (string, error) {
	return "", r.err
}
