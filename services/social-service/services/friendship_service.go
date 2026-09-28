package services

import (
	"context"
	"errors"
	"strings"

	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/repositories"
)

var (
	ErrSelfRequest        = errors.New("cannot send a friendship request to yourself")
	ErrRequestConflict    = errors.New("friendship request already exists")
	ErrBlocked            = errors.New("friendship requests are not allowed between these users")
	ErrRequestNotFound    = errors.New("friendship request not found")
	ErrNotReceiver        = errors.New("only the receiver can manage this request")
	ErrNotSender          = errors.New("only the sender can cancel this request")
	ErrRequestProcessed   = errors.New("friendship request was already processed")
	ErrInvalidUser        = errors.New("authenticated user is required")
	ErrSelfBlock          = errors.New("cannot block yourself")
	ErrAlreadyBlocked     = errors.New("user is already blocked")
	ErrFriendshipNotFound = errors.New("friendship not found")
	ErrNotFriends         = errors.New("users are not friends")
	ErrNotAuthorized      = errors.New("user is not authorized for this operation")
	ErrUserNotFound       = errors.New("user not found")
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
	ListFriends(ctx context.Context, userID string) ([]models.Friend, error)
	RemoveFriend(ctx context.Context, userID, targetID string) error
	GetFriendProfile(ctx context.Context, userID, friendID string) (*models.FriendProfile, error)
	BlockUser(ctx context.Context, blockerID, blockedID string) error
	ValidateInteraction(ctx context.Context, user1ID, user2ID string) error
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

func (s *friendshipService) ListFriends(ctx context.Context, userID string) ([]models.Friend, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidUser
	}
	friends, err := s.repo.ListFriends(ctx, userID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return friends, nil
}

func (s *friendshipService) RemoveFriend(ctx context.Context, userID, targetID string) error {
	userID = strings.TrimSpace(userID)
	targetID = strings.TrimSpace(targetID)
	if userID == "" || targetID == "" {
		return ErrInvalidUser
	}
	err := s.repo.RemoveFriend(ctx, userID, targetID)
	return mapRepositoryError(err)
}

func (s *friendshipService) GetFriendProfile(ctx context.Context, userID, friendID string) (*models.FriendProfile, error) {
	userID = strings.TrimSpace(userID)
	friendID = strings.TrimSpace(friendID)
	if userID == "" || friendID == "" {
		return nil, ErrInvalidUser
	}
	profile, err := s.repo.GetFriendProfile(ctx, userID, friendID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return profile, nil
}

func (s *friendshipService) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	blockerID = strings.TrimSpace(blockerID)
	blockedID = strings.TrimSpace(blockedID)
	if blockerID == "" || blockedID == "" {
		return ErrInvalidUser
	}
	if blockerID == blockedID {
		return ErrSelfBlock
	}
	err := s.repo.BlockUser(ctx, blockerID, blockedID)
	return mapRepositoryError(err)
}

func (s *friendshipService) ValidateInteraction(ctx context.Context, user1ID, user2ID string) error {
	user1ID = strings.TrimSpace(user1ID)
	user2ID = strings.TrimSpace(user2ID)
	if user1ID == "" || user2ID == "" {
		return ErrInvalidUser
	}
	err := s.repo.ValidateInteraction(ctx, user1ID, user2ID)
	return mapRepositoryError(err)
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
	case errors.Is(err, repositories.ErrFriendshipNotFound):
		return ErrFriendshipNotFound
	case errors.Is(err, repositories.ErrSelfBlock):
		return ErrSelfBlock
	case errors.Is(err, repositories.ErrAlreadyBlocked):
		return ErrAlreadyBlocked
	case errors.Is(err, repositories.ErrNotFriends):
		return ErrNotFriends
	case errors.Is(err, repositories.ErrNotAuthorized):
		return ErrNotAuthorized
	case errors.Is(err, repositories.ErrUserNotFound):
		return ErrUserNotFound
	default:
		return err
	}
}
