package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/meloop/auth-service/config"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/repositories"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameExists = errors.New("USERNAME_EXISTS")
	ErrEmailExists    = errors.New("EMAIL_EXISTS")
)

type ValidationError struct {
	Field   string `json:"field"`
	Issue   string `json:"issue"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error)
}

type authService struct {
	config *config.Config
	repo   repositories.UserRepository
}

func NewAuthService(cfg *config.Config, repo repositories.UserRepository) AuthService {
	return &authService{config: cfg, repo: repo}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
	if req.Username == "" {
		return nil, &ValidationError{Field: "username", Issue: "required", Message: "El nombre de usuario es obligatorio"}
	}
	if len(req.Username) > 50 {
		return nil, &ValidationError{Field: "username", Issue: "too_long", Message: "El nombre de usuario no puede superar los 50 caracteres"}
	}

	if req.Email == "" {
		return nil, &ValidationError{Field: "email", Issue: "required", Message: "El correo electrónico es obligatorio"}
	}
	if len(req.Email) > 255 {
		return nil, &ValidationError{Field: "email", Issue: "too_long", Message: "El correo electrónico no puede superar los 255 caracteres"}
	}
	if !emailRegex.MatchString(req.Email) {
		return nil, &ValidationError{Field: "email", Issue: "invalid_format", Message: "El correo electrónico no tiene un formato válido"}
	}

	if req.Password == "" {
		return nil, &ValidationError{Field: "password", Issue: "required", Message: "La contraseña es obligatoria"}
	}
	if len(req.Password) > 72 {
		return nil, &ValidationError{Field: "password", Issue: "too_long", Message: "La contraseña no puede superar los 72 caracteres"}
	}
	if len(req.Password) < s.config.PasswordMinLength {
		return nil, &ValidationError{
			Field:   "password",
			Issue:   "too_short",
			Message: fmt.Sprintf("La contraseña debe tener al menos %d caracteres", s.config.PasswordMinLength),
		}
	}

	existingUser, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameExists
	}

	existingEmail, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingEmail != nil {
		return nil, ErrEmailExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	usuario := &models.Usuario{
		Username:       req.Username,
		Correo:         req.Email,
		ContrasenaHash: string(hashed),
		IDNivel:        nil,
		Experiencia:    0,
	}

	if err := s.repo.Create(ctx, usuario); err != nil {
		if isUniqueViolation(err) {
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "username") {
				return nil, ErrUsernameExists
			}
			if strings.Contains(errStr, "correo") || strings.Contains(errStr, "email") {
				return nil, ErrEmailExists
			}
			return nil, errors.New("el nombre de usuario o correo ya está registrado")
		}
		return nil, err
	}

	return &models.RegisterResponse{
		ID:       usuario.IDUsuario,
		Username: usuario.Username,
		Email:    usuario.Correo,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "23505") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "duplicate key")
}
