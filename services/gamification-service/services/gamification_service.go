package services

import (
	"context"
	"errors"
	"strings"

	"github.com/meloop/gamification-service/models"
	"github.com/meloop/gamification-service/repositories"
)

var (
	ErrInvalidUserID           = errors.New("invalid user id")
	ErrInvalidExperienceAmount = errors.New("experience amount must be greater than zero")
)

// EventPublisher define el contrato para la publicación de eventos en RabbitMQ
type EventPublisher interface {
	PublishLevelUp(ctx context.Context, event models.LevelUpEvent) error
	PublishRewardUnlocked(ctx context.Context, event models.RewardUnlockedEvent) error
}

// GamificationService define las operaciones del servicio de gamificación
type GamificationService interface {
	ProcessExperience(ctx context.Context, userID string, amount int) (*models.ExperienceResult, error)
	GetUserProgress(ctx context.Context, userID string) (*models.UserProgress, error)
}

type gamificationService struct {
	repo      repositories.GamificationRepository
	publisher EventPublisher
}

func NewGamificationService(repo repositories.GamificationRepository, publisher EventPublisher) GamificationService {
	return &gamificationService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *gamificationService) ProcessExperience(ctx context.Context, userID string, amount int) (*models.ExperienceResult, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if amount <= 0 {
		return nil, ErrInvalidExperienceAmount
	}

	// 1. Ejecutar la operación de experiencia dentro de la transacción de base de datos
	result, err := s.repo.AddExperience(ctx, userID, amount)
	if err != nil {
		// Ante cualquier error o rollback en la transacción, NO se publica ningún evento
		return nil, err
	}

	// 2. Transacción confirmada exitosamente. Publicar los eventos correspondientes.
	if s.publisher != nil {
		// Evento level.up: solo si existió una subida de nivel real
		if result.LeveledUp {
			levelUpEvent := models.LevelUpEvent{
				Event:  "level.up",
				UserID: result.UserID,
				Nivel:  result.CurrentLevel,
				Level:  result.CurrentLevel,
			}
			if err := s.publisher.PublishLevelUp(ctx, levelUpEvent); err != nil {
				return result, err
			}
		}

		// Evento reward.unlocked: un evento por cada recompensa incorporada al inventario
		for _, reward := range result.UnlockedRewards {
			rewardEvent := models.RewardUnlockedEvent{
				Event:      "reward.unlocked",
				UserID:     result.UserID,
				RewardID:   reward.IDRecompensa,
				RewardType: reward.Tipo,
				Nivel:      reward.IDNivel,
				Level:      reward.IDNivel,
			}
			if err := s.publisher.PublishRewardUnlocked(ctx, rewardEvent); err != nil {
				return result, err
			}
		}
	}

	return result, nil
}

func (s *gamificationService) GetUserProgress(ctx context.Context, userID string) (*models.UserProgress, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	return s.repo.GetUserProgress(ctx, userID)
}
