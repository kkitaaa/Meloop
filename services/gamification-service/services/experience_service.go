package services

import (
	"context"
	"log"

	"github.com/meloop/gamification-service/repositories"
)

type ExperienceService struct {
	repo   *repositories.ExperienceRepository
	engine *GamificationEngine
}

func NewExperienceService(repo *repositories.ExperienceRepository, engine *GamificationEngine) *ExperienceService {
	return &ExperienceService{repo: repo, engine: engine}
}

func (s *ExperienceService) HandleInteractionEvent(ctx context.Context, event InteractionEvent) error {
	log.Printf("Procesando evento asíncrono para usuario: %s", event.ActorID)

	// 1. Validar que el ID no esté vacío
	if event.ActorID == "" {
		log.Println("⚠️ Advertencia: El ID del usuario llegó vacío. Ignorando evento.")
		return nil
	}

	// 2. Llamar al repositorio para sumar 5 puntos de experiencia
	err := s.repo.AddExperience(ctx, event.ActorID, 5)
	if err != nil {
		log.Printf("❌ Error al sumar experiencia en la BD: %v", err)
		return err
	}

	log.Printf("✅ ¡Experiencia sumada exitosamente al usuario %s en Supabase Local!", event.ActorID)
	return nil
}