package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/repositories"
)

var (
	// ErrUserNotFound indica que el usuario solicitado no existe
	ErrUserNotFound = errors.New("USER_NOT_FOUND")
	// ErrUsernameExists indica que el nombre de usuario ya está registrado por otro usuario (RN-06)
	ErrUsernameExists = errors.New("USERNAME_EXISTS")
	// ErrEmailExists indica que el correo electrónico ya está registrado por otro usuario
	ErrEmailExists = errors.New("EMAIL_EXISTS")
	// ErrEmailSameAsCurrent indica que el nuevo correo es idéntico al actual
	ErrEmailSameAsCurrent = errors.New("EMAIL_SAME_AS_CURRENT")
)

// ValidationError representa un error de validación en un campo específico
type ValidationError struct {
	Field   string `json:"field"`
	Issue   string `json:"issue"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// UserService define la lógica de negocio para la gestión de cuenta de usuario (RF-05)
type UserService interface {
	GetAccount(ctx context.Context, userID string) (*models.AccountResponse, error)
	UpdateUsername(ctx context.Context, userID string, req *models.UpdateUsernameRequest) (*models.AccountResponse, error)
	RequestEmailChange(ctx context.Context, userID string, req *models.RequestEmailChangeRequest) (*models.EmailChangeResponse, error)
}

type userService struct {
	repo repositories.UserRepository
}

// NewUserService crea una nueva instancia del servicio de usuarios
func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// GetAccount obtiene la información básica de la cuenta del usuario autenticado
func (s *userService) GetAccount(ctx context.Context, userID string) (*models.AccountResponse, error) {
	if userID == "" {
		return nil, ErrUserNotFound
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &models.AccountResponse{
		IDUsuario: user.IDUsuario,
		Username:  user.Username,
		Correo:    user.Correo,
	}, nil
}

// UpdateUsername actualiza el nombre de usuario del usuario autenticado aplicando las validaciones de negocio
func (s *userService) UpdateUsername(ctx context.Context, userID string, req *models.UpdateUsernameRequest) (*models.AccountResponse, error) {
	if userID == "" {
		return nil, ErrUserNotFound
	}
	if req == nil || req.Username == "" {
		return nil, &ValidationError{
			Field:   "username",
			Issue:   "required",
			Message: "El nombre de usuario es obligatorio",
		}
	}
	if len(req.Username) > 50 {
		return nil, &ValidationError{
			Field:   "username",
			Issue:   "too_long",
			Message: "El nombre de usuario no puede superar los 50 caracteres",
		}
	}

	currentUser, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, ErrUserNotFound
	}

	// Si el username no cambia, devolvemos los datos actuales
	if currentUser.Username == req.Username {
		return &models.AccountResponse{
			IDUsuario: currentUser.IDUsuario,
			Username:  currentUser.Username,
			Correo:    currentUser.Correo,
		}, nil
	}

	// Comprobar si otro usuario ya tiene ese username registrado
	existing, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.IDUsuario != userID {
		return nil, ErrUsernameExists
	}

	// Persistir la actualización en la base de datos
	updated, err := s.repo.UpdateUsername(ctx, userID, req.Username)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUsernameExists
		}
		return nil, err
	}
	if updated == nil {
		return nil, ErrUserNotFound
	}

	return &models.AccountResponse{
		IDUsuario: updated.IDUsuario,
		Username:  updated.Username,
		Correo:    updated.Correo,
	}, nil
}

// RequestEmailChange procesa la solicitud de cambio de correo respetando las reglas de RN-22
func (s *userService) RequestEmailChange(ctx context.Context, userID string, req *models.RequestEmailChangeRequest) (*models.EmailChangeResponse, error) {
	if userID == "" {
		return nil, ErrUserNotFound
	}
	if req == nil {
		return nil, &ValidationError{
			Field:   "email",
			Issue:   "required",
			Message: "El correo electrónico es obligatorio",
		}
	}

	email := strings.TrimSpace(req.GetEmail())
	if email == "" {
		return nil, &ValidationError{
			Field:   "email",
			Issue:   "required",
			Message: "El correo electrónico es obligatorio",
		}
	}
	if len(email) > 255 {
		return nil, &ValidationError{
			Field:   "email",
			Issue:   "too_long",
			Message: "El correo electrónico no puede superar los 255 caracteres",
		}
	}
	if !emailRegex.MatchString(email) {
		return nil, &ValidationError{
			Field:   "email",
			Issue:   "invalid_format",
			Message: "El correo electrónico no tiene un formato válido",
		}
	}

	currentUser, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, ErrUserNotFound
	}

	// Comprobar que no sea igual al correo actual
	if currentUser.Correo == email {
		return nil, ErrEmailSameAsCurrent
	}

	// Comprobar unicidad del correo electrónico
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.IDUsuario != userID {
		return nil, ErrEmailExists
	}

	// De acuerdo con RN-22 y la infraestructura actual:
	// El correo actual de USUARIO.correo se conserva intacto.
	// La confirmación y reemplazo definitivo del correo dependerá del servicio de envío/verificación de correo.
	return &models.EmailChangeResponse{
		Message:      "Solicitud de cambio de correo recibida. El correo actual se mantendrá hasta su verificación.",
		PendingEmail: email,
	}, nil
}

// isUniqueViolation detecta si el error corresponde a una violación de clave única en PostgreSQL
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "23505") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "duplicate key")
}
