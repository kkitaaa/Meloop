package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/auth-service/config"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/repositories"
)

func TestGetRoleByIDNormalizesAliases(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "usuario", raw: "usuario", want: "usuario"},
		{name: "moderador", raw: "moderador", want: "moderador"},
		{name: "admin", raw: "admin", want: "admin"},
		{name: "administrador", raw: "administrador", want: "admin"},
		{name: "administrator", raw: "ADMINISTRATOR", want: "admin"},
		{name: "moderator", raw: " moderator ", want: "moderador"},
		{name: "desconocido", raw: "root", want: ""},
		{name: "vacio", raw: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{
				users: make(map[string]*models.Usuario),
				getRoleByIDFunc: func(ctx context.Context, id string) (string, error) {
					if id != "user-1" {
						t.Errorf("id = %q, want user-1", id)
					}
					return tt.raw, nil
				},
			}
			sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
			srv := createTestAuthService(&config.Config{}, repo, sessionRepo, nil, nil)

			got, err := srv.GetRoleByID(context.Background(), "user-1")
			if err != nil {
				t.Fatalf("GetRoleByID devolvió error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("rol = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetRoleByIDPropagatesRepositoryError(t *testing.T) {
	repo := &mockUserRepository{
		users: make(map[string]*models.Usuario),
		getRoleByIDFunc: func(ctx context.Context, id string) (string, error) {
			return "", errors.New("postgres no disponible")
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(&config.Config{}, repo, sessionRepo, nil, nil)

	if _, err := srv.GetRoleByID(context.Background(), "user-1"); err == nil {
		t.Fatal("se esperaba error de consulta")
	}
}

func TestGetRoleByIDReturnsUserNotFound(t *testing.T) {
	repo := &mockUserRepository{
		users: make(map[string]*models.Usuario),
		getRoleByIDFunc: func(ctx context.Context, id string) (string, error) {
			return "", repositories.ErrUserNotFound
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(&config.Config{}, repo, sessionRepo, nil, nil)

	_, err := srv.GetRoleByID(context.Background(), "user-1")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Fatalf("error = %v, want ErrUserNotFound", err)
	}
}

func TestGetRoleByIDWithEmptyUserDoesNotQuery(t *testing.T) {
	queried := false
	repo := &mockUserRepository{
		users: make(map[string]*models.Usuario),
		getRoleByIDFunc: func(ctx context.Context, id string) (string, error) {
			queried = true
			return "admin", nil
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(&config.Config{}, repo, sessionRepo, nil, nil)

	got, err := srv.GetRoleByID(context.Background(), "   ")
	if err != nil {
		t.Fatalf("GetRoleByID devolvió error: %v", err)
	}
	if got != "" {
		t.Fatalf("rol = %q, want cadena vacía", got)
	}
	if queried {
		t.Fatal("no se debió consultar el repositorio sin identidad")
	}
}
