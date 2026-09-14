package services

import (
	"context"
	"errors"
	"strings"

	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/repositories"
)

var (
	ErrSelfRequest      = errors.New("cannot send a friendship request to yourself")
	ErrRequestConflict  = errors.New("friendship request already exists")
	ErrBlocked          = errors.New("friendship requests are not allowed between these users")
	ErrRequestNotFound  = errors.New("friendship request not found")
	ErrNotReceiver      = errors.New("only the receiver can manage this request")
	ErrNotSender        = errors.New("only the sender can cancel this request")
	ErrRequestProcessed = errors.New("friendship request was already processed")
	ErrInvalidUser      = errors.New("authenticated user is required")
)

type EventPublisher interface {
	PublishFriendAccepted(ctx context.Context, request *models.FriendRequest) error
}

type FriendshipService interface {
	SendRequest(ctx context.Context, senderID, receiverID string) (*models.FriendRequest, error)
	ReceivedRequests(ctx context.Context, userID string) ([]models.FriendRequest, error)
	SentRequests(ctx context.Context, userID string) ([]models.FriendRequest, error)
	AcceptRequest(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error)
	RejectRequest(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error)
	CancelRequest(ctx context.Context, requestID int, senderID string) (*models.FriendRequest, error)
}

type friendshipService struct {
	repo      repositories.FriendshipRepository
	publisher EventPublisher
}

func NewFriendshipService(repo repositories.FriendshipRepository, publisher EventPublisher) FriendshipService {
	return &friendshipService{repo: repo, publisher: publisher}
}

func (s *friendshipService) SendRequest(ctx context.Context, senderID, receiverID string) (*models.FriendRequest, error) {
	senderID = strings.TrimSpace(senderID)
	receiverID = strings.TrimSpace(receiverID)
	if senderID == "" || receiverID == "" {
		return nil, ErrInvalidUser
	}
	if senderID == receiverID {
		return nil, ErrSelfRequest
	}
	request, err := s.repo.CreatePending(ctx, senderID, receiverID)
	return request, mapRepositoryError(err)
}

func (s *friendshipService) ReceivedRequests(ctx context.Context, userID string) ([]models.FriendRequest, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUser
	}
	return s.repo.ListReceived(ctx, userID)
}

func (s *friendshipService) SentRequests(ctx context.Context, userID string) ([]models.FriendRequest, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUser
	}
	return s.repo.ListSent(ctx, userID)
}

func (s *friendshipService) AcceptRequest(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error) {
	request, err := s.repo.Accept(ctx, requestID, receiverID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	if s.publisher != nil {
		if err := s.publisher.PublishFriendAccepted(ctx, request); err != nil {
			return request, err
		}
	}
	return request, nil
}

func (s *friendshipService) RejectRequest(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error) {
	request, err := s.repo.Reject(ctx, requestID, receiverID)
	return request, mapRepositoryError(err)
}

func (s *friendshipService) CancelRequest(ctx context.Context, requestID int, senderID string) (*models.FriendRequest, error) {
	request, err := s.repo.Cancel(ctx, requestID, senderID)
	return request, mapRepositoryError(err)
}

func mapRepositoryError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repositories.ErrRequestConflict):
		return ErrRequestConflict
	case errors.Is(err, repositories.ErrBlocked):
		return ErrBlocked
	case errors.Is(err, repositories.ErrRequestNotFound):
		return ErrRequestNotFound
	case errors.Is(err, repositories.ErrNotReceiver):
		return ErrNotReceiver
	case errors.Is(err, repositories.ErrNotSender):
		return ErrNotSender
	case errors.Is(err, repositories.ErrRequestProcessed):
		return ErrRequestProcessed
	default:
		return err
	}
}
