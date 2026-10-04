package services

import (
	"context"
	"errors"
	"strings"

	"github.com/meloop/user-service/models"
)

var (
	ErrAdminUserNotFound = errors.New("admin user not found")
	ErrCannotManageSelf  = errors.New("administrator cannot change their own account controls")
)

type AdminRepository interface {
	ListUsers(context.Context) ([]models.AdminUser, error)
	SetSuspension(context.Context, string, bool) (*models.AdminUser, error)
	SetModerator(context.Context, string, bool) (*models.AdminUser, error)
}

type AdminService struct {
	repository AdminRepository
}

func NewAdminService(repository AdminRepository) *AdminService {
	return &AdminService{repository: repository}
}

func (s *AdminService) ListUsers(ctx context.Context) ([]models.AdminUser, error) {
	return s.repository.ListUsers(ctx)
}

func (s *AdminService) SetSuspension(ctx context.Context, actorID, targetID string, suspended bool) (*models.AdminUser, error) {
	if err := validateAdminTarget(actorID, targetID); err != nil {
		return nil, err
	}
	user, err := s.repository.SetSuspension(ctx, targetID, suspended)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrAdminUserNotFound
	}
	return user, nil
}

func (s *AdminService) SetModerator(ctx context.Context, actorID, targetID string, enabled bool) (*models.AdminUser, error) {
	if err := validateAdminTarget(actorID, targetID); err != nil {
		return nil, err
	}
	user, err := s.repository.SetModerator(ctx, targetID, enabled)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrAdminUserNotFound
	}
	return user, nil
}

func validateAdminTarget(actorID, targetID string) error {
	if strings.TrimSpace(targetID) == "" {
		return ErrAdminUserNotFound
	}
	if actorID == targetID {
		return ErrCannotManageSelf
	}
	return nil
}
