package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	ErrUsernameExists     = errors.New("USERNAME_EXISTS")
	ErrEmailExists        = errors.New("EMAIL_EXISTS")
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")
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
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*models.SessionUser, error)
	ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error
}

type authService struct {
	config      *config.Config
	repo        repositories.UserRepository
	sessionRepo repositories.SessionRepository
}

func NewAuthService(
	cfg *config.Config,
	repo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
) AuthService {
	return &authService{
		config:      cfg,
		repo:        repo,
		sessionRepo: sessionRepo,
	}
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

func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	if req.Email == "" {
		return nil, &ValidationError{Field: "email", Issue: "required", Message: "El correo electrónico es obligatorio"}
	}
	if req.Password == "" {
		return nil, &ValidationError{Field: "password", Issue: "required", Message: "La contraseña es obligatoria"}
	}

	usuario, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if usuario == nil {
		// Run dummy comparison to prevent timing attacks
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$dummyhashplaceholderforsecurityreasonsinfo"), []byte(req.Password))
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(usuario.ContrasenaHash), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	sessionUser := &models.SessionUser{
		ID:       usuario.IDUsuario,
		Username: usuario.Username,
		Email:    usuario.Correo,
	}

	err = s.sessionRepo.Create(ctx, token, sessionUser, s.config.SessionTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to save session in redis: %w", err)
	}

	expiresInSeconds := int(s.config.SessionTTL.Seconds())

	return &models.LoginResponse{
		SessionToken: token,
		ExpiresIn:    expiresInSeconds,
		User:         *sessionUser,
	}, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessionRepo.Delete(ctx, token)
}

func (s *authService) ValidateSession(ctx context.Context, token string) (*models.SessionUser, error) {
	if token == "" {
		return nil, errors.New("empty session token")
	}
	user, err := s.sessionRepo.Get(ctx, token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("session invalid or expired")
	}
	return user, nil
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ChangePassword actualiza la contraseña del usuario tras validar las credenciales actuales
func (s *authService) ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error {
	if userID == "" {
		return ErrInvalidCredentials
	}
	if req == nil {
		return &ValidationError{
			Field:   "current_password",
			Issue:   "required",
			Message: "La contraseña actual es obligatoria",
		}
	}
	if req.CurrentPassword == "" {
		return &ValidationError{
			Field:   "current_password",
			Issue:   "required",
			Message: "La contraseña actual es obligatoria",
		}
	}
	if req.NewPassword == "" {
		return &ValidationError{
			Field:   "new_password",
			Issue:   "required",
			Message: "La nueva contraseña es obligatoria",
		}
	}
	if len(req.NewPassword) < s.config.PasswordMinLength {
		return &ValidationError{
			Field:   "new_password",
			Issue:   "too_short",
			Message: fmt.Sprintf("La nueva contraseña debe tener al menos %d caracteres", s.config.PasswordMinLength),
		}
	}
	if len(req.NewPassword) > 72 {
		return &ValidationError{
			Field:   "new_password",
			Issue:   "too_long",
			Message: "La nueva contraseña no puede superar los 72 caracteres",
		}
	}
	if req.NewPassword == req.CurrentPassword {
		return &ValidationError{
			Field:   "new_password",
			Issue:   "same_as_current",
			Message: "La nueva contraseña no puede ser igual a la contraseña actual",
		}
	}

	usuario, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if usuario == nil {
		return ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(usuario.ContrasenaHash), []byte(req.CurrentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("fallo al generar hash de contraseña: %w", err)
	}

	return s.repo.UpdatePassword(ctx, userID, string(hashed))
}
