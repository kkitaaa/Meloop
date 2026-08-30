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
	"golang.org/x/crypto/bcrypt"
	"github.com/meloop/auth-service/config"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/repositories"
)

var (
	ErrUsernameExists     = errors.New("USERNAME_EXISTS")
	ErrEmailExists        = errors.New("EMAIL_EXISTS")
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")
)

// ValidationError represents validation error details for response mapping
type ValidationError struct {
	Field   string `json:"field"`
	Issue   string `json:"issue"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// AuthService defines authentication and session business logic operations
type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*models.SessionUser, error)
}

type authService struct {
	config      *config.Config
	repo        repositories.AccountRepository
	sessionRepo repositories.SessionRepository
}

// NewAuthService creates a new AuthService implementation with session support
func NewAuthService(
	cfg *config.Config,
	repo repositories.AccountRepository,
	sessionRepo repositories.SessionRepository,
) AuthService {
	return &authService{
		config:      cfg,
		repo:        repo,
		sessionRepo: sessionRepo,
	}
}

// Simple and standard email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
	// 1. Mandatory fields checks & defensive length validation (username)
	if req.Username == "" {
		return nil, &ValidationError{Field: "username", Issue: "required", Message: "El nombre de usuario es obligatorio"}
	}
	if len(req.Username) > 50 {
		return nil, &ValidationError{Field: "username", Issue: "too_long", Message: "El nombre de usuario no puede superar los 50 caracteres"}
	}

	// 2. Mandatory fields checks & format checks & defensive length validation (email)
	if req.Email == "" {
		return nil, &ValidationError{Field: "email", Issue: "required", Message: "El correo electrónico es obligatorio"}
	}
	if len(req.Email) > 255 {
		return nil, &ValidationError{Field: "email", Issue: "too_long", Message: "El correo electrónico no puede superar los 255 caracteres"}
	}
	if !emailRegex.MatchString(req.Email) {
		return nil, &ValidationError{Field: "email", Issue: "invalid_format", Message: "El correo electrónico no tiene un formato válido"}
	}

	// 3. Mandatory fields checks & password strength & defensive length validation (password)
	if req.Password == "" {
		return nil, &ValidationError{Field: "password", Issue: "required", Message: "La contraseña es obligatoria"}
	}
	if len(req.Password) > 72 {
		// Enforce maximum length of 72 characters because bcrypt truncates inputs longer than 72 bytes
		return nil, &ValidationError{Field: "password", Issue: "too_long", Message: "La contraseña no puede superar los 72 caracteres"}
	}
	if len(req.Password) < s.config.PasswordMinLength {
		return nil, &ValidationError{
			Field:   "password",
			Issue:   "too_short",
			Message: fmt.Sprintf("La contraseña debe tener al menos %d caracteres", s.config.PasswordMinLength),
		}
	}

	// 4. Uniqueness check for username
	existingUser, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameExists
	}

	// 5. Uniqueness check for email
	existingEmail, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingEmail != nil {
		return nil, ErrEmailExists
	}

	// 6. Secure password hashing with bcrypt
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	account := &models.Account{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashed),
	}

	// 7. Persist database record
	if err := s.repo.Create(ctx, account); err != nil {
		// Map database unique constraint violation errors in case of race conditions
		if isUniqueViolation(err) {
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "username") {
				return nil, ErrUsernameExists
			}
			if strings.Contains(errStr, "email") {
				return nil, ErrEmailExists
			}
			return nil, errors.New("el nombre de usuario o correo ya está registrado")
		}
		return nil, err
	}

	// 8. Return response omitting credentials
	return &models.RegisterResponse{
		ID:        account.ID,
		Username:  account.Username,
		Email:     account.Email,
		CreatedAt: account.CreatedAt,
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

	account, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if account == nil {
		// Run dummy comparison to prevent timing attacks
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$dummyhashplaceholderforsecurityreasonsinfo"), []byte(req.Password))
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	sessionUser := &models.SessionUser{
		ID:       account.ID,
		Username: account.Username,
		Email:    account.Email,
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
