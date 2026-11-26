package services

import (
	"context"
	"errors"
	"strings"

	"github.com/meloop/notification-service/models"
	"github.com/meloop/notification-service/repositories"
)

type InboxRepository interface {
	ListByUserID(context.Context, string, int, int) ([]models.NotificationResponse, error)
	MarkRead(context.Context, string, string) error
	MarkAllRead(context.Context, string) (int64, error)
}

type InboxService struct {
	repository InboxRepository
}

func NewInboxService(repository InboxRepository) *InboxService {
	return &InboxService{repository: repository}
}

func (s *InboxService) List(ctx context.Context, userID string, limit, offset int) ([]models.NotificationResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	return s.repository.ListByUserID(ctx, userID, limit, offset)
}

func (s *InboxService) MarkRead(ctx context.Context, userID, notificationID string) error {
	userID = strings.TrimSpace(userID)
	notificationID = strings.TrimSpace(notificationID)
	if userID == "" || notificationID == "" {
		return repositories.ErrNotificationNotFound
	}
	return s.repository.MarkRead(ctx, userID, notificationID)
}

func (s *InboxService) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return 0, errors.New("user ID is required")
	}
	return s.repository.MarkAllRead(ctx, userID)
}
