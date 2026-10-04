package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/user-service/models"
)

type adminRepositoryFake struct {
	user *models.AdminUser
}

func (repository *adminRepositoryFake) ListUsers(context.Context) ([]models.AdminUser, error) {
	return []models.AdminUser{}, nil
}

func (repository *adminRepositoryFake) SetSuspension(_ context.Context, userID string, suspended bool) (*models.AdminUser, error) {
	return &models.AdminUser{IDUsuario: userID, Suspendido: suspended}, nil
}

func (repository *adminRepositoryFake) SetModerator(_ context.Context, userID string, enabled bool) (*models.AdminUser, error) {
	return &models.AdminUser{IDUsuario: userID, PuedeModerar: enabled}, nil
}

func TestAdminServicePreventsSelfSuspension(t *testing.T) {
	service := NewAdminService(&adminRepositoryFake{})
	_, err := service.SetSuspension(context.Background(), "admin-1", "admin-1", true)
	if !errors.Is(err, ErrCannotManageSelf) {
		t.Fatalf("SetSuspension() error = %v, want ErrCannotManageSelf", err)
	}
}

func TestAdminServicePreventsSelfModeratorChange(t *testing.T) {
	service := NewAdminService(&adminRepositoryFake{})
	_, err := service.SetModerator(context.Background(), "admin-1", "admin-1", true)
	if !errors.Is(err, ErrCannotManageSelf) {
		t.Fatalf("SetModerator() error = %v, want ErrCannotManageSelf", err)
	}
}
