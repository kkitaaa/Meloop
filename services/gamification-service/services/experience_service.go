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

// HandleInteractionEvent es el punto de entrada cuando llega un evento (ej. desde RabbitMQ)
func (s *ExperienceService) HandleInteractionEvent(ctx context.Context, event InteractionEvent) error {
	// 1. Pasar el evento por los filtros anti-abuso y calcular XP[cite: 8]
	xp, err := s.engine.ProcessEvent(event)
	if err != nil {
		// Logueamos el rechazo (rate limit o auto-interacción) pero no devolvemos error fatal
		// para que el sistema de mensajería no reintente procesar un evento abusivo.
		log.Printf("Evento descartado por filtros anti-abuso: %v", err)
		return nil
	}

	// 2. Si pasa las validaciones, asignar la experiencia exacta al usuario[cite: 8]
	log.Printf("Evento válido. Asignando +%d XP al usuario %s", xp, event.ActorID)

	// Aquí llamaríamos a la función atómica que construimos anteriormente
	err = s.repo.AddExperience(ctx, event.ActorID, xp)
	if err != nil {
		return fmt.Errorf("error al guardar experiencia en BD: %w", err)
	}

	return nil
}
