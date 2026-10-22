package services

import (
	"context"
	"fmt"
	"log"

	"github.com/meloop/gamification-service/repositories"
)

type ExperienceService struct {
	repo   *repositories.ExperienceRepository
	engine *GamificationEngine
}

func NewExperienceService(repo *repositories.ExperienceRepository, engine *GamificationEngine) *ExperienceService {
	return &ExperienceService{
		repo:   repo,
		engine: engine,
	}
}

// AddUserExperience maneja la lógica de negocio y define las reglas de tolerancia a fallos
func (s *ExperienceService) AddUserExperience(ctx context.Context, userID string, xpToAdd int) (int, error) {
	// Definimos el número máximo de reintentos permitidos si hay fallos temporales en Supabase
	maxRetries := 3

	// Ejecutamos la suma atómica con la protección de reintentos
	newXP, err := s.repo.AddExperienceWithRetry(ctx, userID, xpToAdd, maxRetries)
	if err != nil {
		return 0, err
	}

	return newXP, nil
}

// HandleInteractionEvent es el punto de entrada cuando llega un evento (ej. desde RabbitMQ)
func (s *ExperienceService) HandleInteractionEvent(ctx context.Context, event InteractionEvent) error {
	// 1. Pasar el evento por los filtros anti-abuso y calcular XP
	xp, err := s.engine.ProcessEvent(event)
	if err != nil {
		log.Printf("Evento descartado por filtros anti-abuso: %v", err)
		return nil
	}

	// 2. Si pasa las validaciones, asignar la experiencia exacta al usuario
	log.Printf("Evento válido. Asignando +%d XP al usuario %s", xp, event.ActorID)

	err = s.repo.AddExperience(ctx, event.ActorID, xp)
	if err != nil {
		return fmt.Errorf("error al guardar experiencia en BD: %w", err)
	}

	return nil
}
