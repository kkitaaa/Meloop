package services

import (
	"context"
	"github.com/meloop/gamification-service/repositories"
)

type ExperienceService struct {
	repo *repositories.PostgresExperienceRepository
}

// NewExperienceService inicializa la capa de servicio
func NewExperienceService(repo *repositories.PostgresExperienceRepository) *ExperienceService {
	return &ExperienceService{repo: repo}
}

// AddUserExperience maneja la lógica de negocio y define las reglas de tolerancia a fallos
func (s *ExperienceService) AddUserExperience(ctx context.Context, userID string, xpToAdd int) (int, error) {
	// Definimos el número máximo de reintentos permitidos si hay fallos temporales en Supabase
	maxRetries := 3

	// Ejecutamos la suma atómica con la protección de reintentos
	newXP, err := s.repo.AddExperienceWithRetry(ctx, userID, xpToAdd, maxRetries)
	if err != nil {
		// Aquí podrías emitir métricas o logs en un futuro
		return 0, err
	}

	return newXP, nil
}