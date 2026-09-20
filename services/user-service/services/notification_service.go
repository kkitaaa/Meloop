package services

import (
	"context"
	"strings"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/repositories"
)

// NotificationConfigService define la lógica de negocio para la configuración de notificaciones (RF-05 / RF-52)
type NotificationConfigService interface {
	GetSettings(ctx context.Context, userID string) ([]models.NotificationSettingResponse, error)
	UpdateSetting(ctx context.Context, userID string, tipo string, habilitada bool) (*models.NotificationSettingResponse, error)
	UpdateBatch(ctx context.Context, userID string, items []models.NotificationSettingItem) ([]models.NotificationSettingResponse, error)
}

type notificationConfigService struct {
	notifRepo repositories.NotificationConfigRepository
	userRepo  repositories.UserRepository
}

// NewNotificationConfigService crea una nueva instancia de NotificationConfigService
func NewNotificationConfigService(notifRepo repositories.NotificationConfigRepository, userRepo repositories.UserRepository) NotificationConfigService {
	return &notificationConfigService{
		notifRepo: notifRepo,
		userRepo:  userRepo,
	}
}

// GetSettings consulta las configuraciones de notificación del usuario autenticado
func (s *notificationConfigService) GetSettings(ctx context.Context, userID string) ([]models.NotificationSettingResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	configs, err := s.notifRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toNotificationSettingResponses(configs), nil
}

// UpdateSetting actualiza el estado de habilitación de un único tipo de notificación
func (s *notificationConfigService) UpdateSetting(ctx context.Context, userID string, tipo string, habilitada bool) (*models.NotificationSettingResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	tipoNorm := strings.ToUpper(strings.TrimSpace(tipo))
	if !models.IsValidNotificationType(tipoNorm) {
		return nil, &ValidationError{
			Field:   "tipo_notificacion",
			Issue:   "invalid_type",
			Message: "El tipo de notificación especificado no es válido",
		}
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	updated, err := s.notifRepo.Upsert(ctx, userID, tipoNorm, habilitada)
	if err != nil {
		return nil, err
	}

	return &models.NotificationSettingResponse{
		TipoNotificacion: updated.TipoNotificacion,
		Habilitada:       updated.Habilitada,
	}, nil
}

// UpdateBatch actualiza múltiples tipos de notificación en una sola operación
func (s *notificationConfigService) UpdateBatch(ctx context.Context, userID string, items []models.NotificationSettingItem) ([]models.NotificationSettingResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	if len(items) == 0 {
		return nil, &ValidationError{
			Field:   "configuraciones",
			Issue:   "empty",
			Message: "Debe proporcionar al menos una configuración para actualizar",
		}
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Validar todos los elementos antes de ejecutar la persistencia
	var normalizedItems []models.NotificationSettingItem
	for _, item := range items {
		tipoNorm := strings.ToUpper(strings.TrimSpace(item.TipoNotificacion))
		if !models.IsValidNotificationType(tipoNorm) {
			return nil, &ValidationError{
				Field:   "tipo_notificacion",
				Issue:   "invalid_type",
				Message: "Tipo de notificación no válido: " + item.TipoNotificacion,
			}
		}
		if item.Habilitada == nil {
			return nil, &ValidationError{
				Field:   "habilitada",
				Issue:   "required",
				Message: "El campo habilitada es obligatorio para " + item.TipoNotificacion,
			}
		}
		normalizedItems = append(normalizedItems, models.NotificationSettingItem{
			TipoNotificacion: tipoNorm,
			Habilitada:       item.Habilitada,
		})
	}

	configs, err := s.notifRepo.UpsertBatch(ctx, userID, normalizedItems)
	if err != nil {
		return nil, err
	}

	return toNotificationSettingResponses(configs), nil
}

func toNotificationSettingResponses(configs []models.ConfiguracionNotificacion) []models.NotificationSettingResponse {
	resp := make([]models.NotificationSettingResponse, 0, len(configs))
	for _, c := range configs {
		resp = append(resp, models.NotificationSettingResponse{
			TipoNotificacion: c.TipoNotificacion,
			Habilitada:       c.Habilitada,
		})
	}
	return resp
}
