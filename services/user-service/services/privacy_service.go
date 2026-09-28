package services

import (
	"context"
	"strings"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/repositories"
)

// PrivacyService define los casos de uso para la gestión de configuración de privacidad (RF-05 / RF-55)
type PrivacyService interface {
	GetPrivacy(ctx context.Context, userID string) (*models.PrivacyResponse, error)
	UpdatePrivacy(ctx context.Context, userID string, req *models.UpdatePrivacyRequest) (*models.PrivacyResponse, error)
}

type privacyService struct {
	privacyRepo repositories.PrivacyRepository
	userRepo    repositories.UserRepository
}

// NewPrivacyService crea una nueva instancia de PrivacyService
func NewPrivacyService(privacyRepo repositories.PrivacyRepository, userRepo repositories.UserRepository) PrivacyService {
	return &privacyService{
		privacyRepo: privacyRepo,
		userRepo:    userRepo,
	}
}

// GetPrivacy obtiene la configuración de privacidad del usuario autenticado
func (s *privacyService) GetPrivacy(ctx context.Context, userID string) (*models.PrivacyResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	// Verificar existencia del usuario
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	conf, err := s.privacyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		return nil, ErrUserNotFound
	}

	return toPrivacyResponse(conf), nil
}

// UpdatePrivacy actualiza de forma parcial o total la configuración de privacidad del usuario
func (s *privacyService) UpdatePrivacy(ctx context.Context, userID string, req *models.UpdatePrivacyRequest) (*models.PrivacyResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	if req == nil {
		return nil, &ValidationError{
			Field:   "body",
			Issue:   "required",
			Message: "El cuerpo de la petición no puede estar vacío",
		}
	}

	// Verificar existencia del usuario
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Validar valores enviados
	if req.VisibilidadPerfil != nil {
		val := strings.ToUpper(strings.TrimSpace(*req.VisibilidadPerfil))
		if !isValidVisibility(val) {
			return nil, &ValidationError{
				Field:   "visibilidad_perfil",
				Issue:   "invalid_value",
				Message: "Valor no permitido para visibilidad_perfil. Valores válidos: PUBLICO, AMIGOS, PRIVADO",
			}
		}
		*req.VisibilidadPerfil = val
	}

	if req.VisibilidadPublicaciones != nil {
		val := strings.ToUpper(strings.TrimSpace(*req.VisibilidadPublicaciones))
		if !isValidVisibility(val) {
			return nil, &ValidationError{
				Field:   "visibilidad_publicaciones",
				Issue:   "invalid_value",
				Message: "Valor no permitido para visibilidad_publicaciones. Valores válidos: PUBLICO, AMIGOS, PRIVADO",
			}
		}
		*req.VisibilidadPublicaciones = val
	}

	if req.VisibilidadInteracciones != nil {
		val := strings.ToUpper(strings.TrimSpace(*req.VisibilidadInteracciones))
		if !isValidVisibility(val) {
			return nil, &ValidationError{
				Field:   "visibilidad_interacciones",
				Issue:   "invalid_value",
				Message: "Valor no permitido para visibilidad_interacciones. Valores válidos: PUBLICO, AMIGOS, PRIVADO",
			}
		}
		*req.VisibilidadInteracciones = val
	}

	if req.RecepcionMensajes != nil {
		val := strings.ToUpper(strings.TrimSpace(*req.RecepcionMensajes))
		if !isValidReception(val) {
			return nil, &ValidationError{
				Field:   "recepcion_mensajes",
				Issue:   "invalid_value",
				Message: "Valor no permitido para recepcion_mensajes. Valores válidos: TODOS, AMIGOS, NADIE",
			}
		}
		*req.RecepcionMensajes = val
	}

	recepcionSolicitudes := req.GetRecepcionSolicitudes()
	if recepcionSolicitudes != nil {
		val := strings.ToUpper(strings.TrimSpace(*recepcionSolicitudes))
		if !isValidReception(val) {
			return nil, &ValidationError{
				Field:   "recepcion_solicitudes_amistad",
				Issue:   "invalid_value",
				Message: "Valor no permitido para recepcion_solicitudes_amistad. Valores válidos: TODOS, AMIGOS, NADIE",
			}
		}
		*recepcionSolicitudes = val
	}

	// Obtener la configuración actual para realizar la actualización parcial
	current, err := s.privacyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		current = &models.ConfiguracionPrivacidad{
			IDUsuario:                   userID,
			VisibilidadPerfil:           models.VisibilidadPublico,
			VisibilidadPublicaciones:    models.VisibilidadPublico,
			VisibilidadInteracciones:    models.VisibilidadPublico,
			RecepcionMensajes:           models.RecepcionTodos,
			RecepcionSolicitudesAmistad: models.RecepcionTodos,
		}
	}

	// Aplicar únicamente los campos proporcionados
	if req.VisibilidadPerfil != nil {
		current.VisibilidadPerfil = *req.VisibilidadPerfil
	}
	if req.VisibilidadPublicaciones != nil {
		current.VisibilidadPublicaciones = *req.VisibilidadPublicaciones
	}
	if req.VisibilidadInteracciones != nil {
		current.VisibilidadInteracciones = *req.VisibilidadInteracciones
	}
	if req.RecepcionMensajes != nil {
		current.RecepcionMensajes = *req.RecepcionMensajes
	}
	if recepcionSolicitudes != nil {
		current.RecepcionSolicitudesAmistad = *recepcionSolicitudes
	}

	// Persistir cambios
	updated, err := s.privacyRepo.Update(ctx, current)
	if err != nil {
		return nil, err
	}

	return toPrivacyResponse(updated), nil
}

func isValidVisibility(v string) bool {
	return v == models.VisibilidadPublico || v == models.VisibilidadAmigos || v == models.VisibilidadPrivado
}

func isValidReception(r string) bool {
	return r == models.RecepcionTodos || r == models.RecepcionAmigos || r == models.RecepcionNadie
}

func toPrivacyResponse(c *models.ConfiguracionPrivacidad) *models.PrivacyResponse {
	return &models.PrivacyResponse{
		IDUsuario:                   c.IDUsuario,
		VisibilidadPerfil:           c.VisibilidadPerfil,
		VisibilidadPublicaciones:    c.VisibilidadPublicaciones,
		VisibilidadInteracciones:    c.VisibilidadInteracciones,
		RecepcionMensajes:           c.RecepcionMensajes,
		RecepcionSolicitudesAmistad: c.RecepcionSolicitudesAmistad,
	}
}
