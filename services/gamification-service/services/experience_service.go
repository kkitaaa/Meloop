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
	return nil
}