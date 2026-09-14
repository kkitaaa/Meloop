package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/repositories"
)

type friendshipRepositoryFake struct {
	createErr error
	acceptErr error
	rejectErr error
	cancelErr error
	request   *models.FriendRequest
	created   bool
}

func (f *friendshipRepositoryFake) CreatePending(context.Context, string, string) (*models.FriendRequest, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.created = true
	return f.request, nil
}

func (f *friendshipRepositoryFake) ListReceived(context.Context, string) ([]models.FriendRequest, error) {
	return []models.FriendRequest{}, nil
}

func (f *friendshipRepositoryFake) ListSent(context.Context, string) ([]models.FriendRequest, error) {
	return []models.FriendRequest{}, nil
}

func (f *friendshipRepositoryFake) Accept(context.Context, int, string) (*models.FriendRequest, error) {
	if f.acceptErr != nil {
		return nil, f.acceptErr
	}
	return f.request, nil
}

func (f *friendshipRepositoryFake) Reject(context.Context, int, string) (*models.FriendRequest, error) {
	if f.rejectErr != nil {
		return nil, f.rejectErr
	}
	return f.request, nil
}

func (f *friendshipRepositoryFake) Cancel(context.Context, int, string) (*models.FriendRequest, error) {
	if f.cancelErr != nil {
		return nil, f.cancelErr
	}
	return f.request, nil
}

type eventPublisherFake struct {
	called bool
}

func (f *eventPublisherFake) PublishFriendAccepted(context.Context, *models.FriendRequest) error {
	f.called = true
	return nil
}

func TestFriendshipServiceSendRequest(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "valid request"},
		{name: "self request", err: ErrSelfRequest, want: ErrSelfRequest},
		{name: "duplicate request", err: repositories.ErrRequestConflict, want: ErrRequestConflict},
		{name: "inverse pending request", err: repositories.ErrRequestConflict, want: ErrRequestConflict},
		{name: "blocked users", err: repositories.ErrBlocked, want: ErrBlocked},
		{name: "accepted friendship", err: repositories.ErrRequestConflict, want: ErrRequestConflict},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &friendshipRepositoryFake{
				createErr: test.err,
				request:   &models.FriendRequest{ID: 1, SenderID: "a", ReceiverID: "b", Status: models.FriendshipPending},
			}
			service := NewFriendshipService(repo, nil)
			_, err := service.SendRequest(context.Background(), "a", map[bool]string{true: "a", false: "b"}[test.name == "self request"])
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if test.want == nil && !repo.created {
				t.Fatal("expected repository to create the request")
			}
		})
	}
}

func TestFriendshipServiceActionsEnforceAuthorizationAndState(t *testing.T) {
	tests := []struct {
		name string
		err  error
		call func(FriendshipService) error
	}{
		{name: "accept valid", call: func(s FriendshipService) error {
			_, err := s.AcceptRequest(context.Background(), 1, "receiver")
			return err
		}},
		{name: "reject valid", call: func(s FriendshipService) error {
			_, err := s.RejectRequest(context.Background(), 1, "receiver")
			return err
		}},
		{name: "cancel valid", call: func(s FriendshipService) error {
			_, err := s.CancelRequest(context.Background(), 1, "sender")
			return err
		}},
		{name: "accept non receiver", err: ErrNotReceiver, call: func(s FriendshipService) error {
			_, err := s.AcceptRequest(context.Background(), 1, "other")
			return err
		}},
		{name: "reject non receiver", err: ErrNotReceiver, call: func(s FriendshipService) error {
			_, err := s.RejectRequest(context.Background(), 1, "other")
			return err
		}},
		{name: "cancel non sender", err: ErrNotSender, call: func(s FriendshipService) error {
			_, err := s.CancelRequest(context.Background(), 1, "other")
			return err
		}},
		{name: "missing request", err: ErrRequestNotFound, call: func(s FriendshipService) error {
			_, err := s.AcceptRequest(context.Background(), 99, "receiver")
			return err
		}},
		{name: "already processed", err: ErrRequestProcessed, call: func(s FriendshipService) error {
			_, err := s.RejectRequest(context.Background(), 1, "receiver")
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &friendshipRepositoryFake{
				request: &models.FriendRequest{ID: 1, SenderID: "sender", ReceiverID: "receiver", Status: models.FriendshipAccepted},
			}
			switch test.name {
			case "accept non receiver", "reject non receiver":
				repo.acceptErr = repositories.ErrNotReceiver
				repo.rejectErr = repositories.ErrNotReceiver
			case "cancel non sender":
				repo.cancelErr = repositories.ErrNotSender
			case "missing request":
				repo.acceptErr = repositories.ErrRequestNotFound
			case "already processed":
				repo.rejectErr = repositories.ErrRequestProcessed
			}
			if err := test.call(NewFriendshipService(repo, nil)); !errors.Is(err, test.err) {
				t.Fatalf("error = %v, want %v", err, test.err)
			}
		})
	}
}

func TestFriendshipServiceAcceptPublishesEventAfterRepositoryAction(t *testing.T) {
	repo := &friendshipRepositoryFake{
		request: &models.FriendRequest{ID: 1, SenderID: "sender", ReceiverID: "receiver", Status: models.FriendshipAccepted},
	}
	publisher := &eventPublisherFake{}
	service := NewFriendshipService(repo, publisher)

	request, err := service.AcceptRequest(context.Background(), 1, "receiver")
	if err != nil {
		t.Fatalf("AcceptRequest() error = %v", err)
	}
	if request.Status != models.FriendshipAccepted {
		t.Fatalf("status = %q, want %q", request.Status, models.FriendshipAccepted)
	}
	if !publisher.called {
		t.Fatal("expected friend.accepted event to be published")
	}
}
