package services

import (
	"context"
	"fmt"
	"log"

	"github.com/meloop/gamification-service/repositories"
)

type ExperienceService struct {
	repo   *repositories.PostgresExperienceRepository
	engine *GamificationEngine
}

// NewExperienceService inicializa la capa de servicio
func NewExperienceService(repo *repositories.PostgresExperienceRepository, engine *GamificationEngine) *ExperienceService {
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
		// Aquí podrías emitir métricas o logs en un futuro
		return 0, err
	}

	return newXP, nil
}

// HandleInteractionEvent es el punto de entrada cuando llega un evento (ej. desde RabbitMQ)
func (s *ExperienceService) HandleInteractionEvent(ctx context.Context, event InteractionEvent) error {
	// 1. Pasar el evento por los filtros anti-abuso y calcular XP
	xp, err := s.engine.ProcessEvent(event)
	if err != nil {
		// Logueamos el rechazo (rate limit o auto-interacción) pero no devolvemos error fatal
		// para que el sistema de mensajería no reintente procesar un evento abusivo.
		log.Printf("Evento descartado por filtros anti-abuso: %v", err)
		return nil
	}

	// 2. Si pasa las validaciones, asignar la experiencia exacta al usuario
	log.Printf("Evento válido. Asignando +%d XP al usuario %s", xp, event.ActorID)

	// Llamamos a nuestra propia función de servicio que ya incluye la lógica de reintentos
	_, err = s.AddUserExperience(ctx, event.ActorID, xp)
	if err != nil {
		return fmt.Errorf("error al guardar experiencia en BD: %w", err)
	}

	return nil
}