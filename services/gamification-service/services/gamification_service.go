package services

import (
	"log"

	"github.com/meloop/gamification-service/models"
	"github.com/meloop/gamification-service/repositories"
)

type GamificationService struct {
	repo *repositories.GamificationRepository
}

func NewGamificationService(repo *repositories.GamificationRepository) *GamificationService {
	return &GamificationService{repo: repo}
}

// AddXP suma experiencia, recalcula el nivel y desbloquea beneficios de forma dinámica
func (s *GamificationService) AddXP(userID string, xpToAdd int) (models.UserProgress, []string, error) {
	progress, err := s.repo.GetOrCreateProgress(userID)
	if err != nil {
		return progress, nil, err
	}

	progress.TotalXP += xpToAdd
	var unlockedRewards []string

	// Usamos un bucle por si ganó mucha XP de golpe y sube varios niveles a la vez
	for {
		nextLevel := progress.Level + 1
		rule, err := s.repo.GetLevelRule(nextLevel)
		
		// Si no encuentra regla para el nivel (ej: llegó al nivel máximo), rompemos el bucle
		if err != nil {
			break
		}

		// Verificamos el cruce de umbral matemático
		if progress.TotalXP >= rule.RequiredXP {
			progress.Level = nextLevel
			log.Printf("Usuario %s alcanzó el nivel %d", userID, nextLevel)

			// Si este nivel tiene una recompensa estética asociada, se la habilitamos
			if rule.RewardName != "" {
				reward := models.UserReward{
					UserID:     userID,
					RewardName: rule.RewardName,
				}
				s.repo.UnlockReward(reward)
				unlockedRewards = append(unlockedRewards, rule.RewardName)
				log.Printf("Desbloqueo de beneficio: %s", rule.RewardName)
			}
		} else {
			// Si la XP no alcanza para el siguiente nivel, paramos de calcular
			break
		}
	}

	// Guardamos el nuevo estado transaccional
	err = s.repo.UpdateProgress(progress)
	return progress, unlockedRewards, err
}