package services

import (
	"math"

	"github.com/meloop/gamification-service/models"
	"github.com/meloop/gamification-service/repositories"
)

type ProgressService struct {
	repo repositories.ProgressRepository
}

func NewProgressService(repo repositories.ProgressRepository) *ProgressService {
	return &ProgressService{repo: repo}
}

func (s *ProgressService) GetUserProgress(userID string) (models.ProgressResponse, error) {
	// 1. Obtener datos base (nivel y exp)
	baseData, err := s.repo.GetBaseProgress(userID)
	if err != nil {
		return models.ProgressResponse{}, err
	}

	// 2. Obtener recompensas desbloqueadas
	unlocked, err := s.repo.GetUnlockedRewards(userID)
	if err != nil || len(unlocked) == 0 {
		unlocked = []models.RewardInfo{} // Aseguramos arreglo vacío en vez de null
	}

	// 3. Obtener próximas recompensas
	nextRewards, err := s.repo.GetNextRewards(baseData.NivelActual)
	if err != nil || len(nextRewards) == 0 {
		nextRewards = []models.RewardInfo{}
	}

	// 4. Calcular porcentaje de avance (protegido contra división por cero)
	var porcentaje float64
	if baseData.ExperienciaProxima > 0 {
		porcentaje = (float64(baseData.ExperienciaActual) / float64(baseData.ExperienciaProxima)) * 100.0
		porcentaje = math.Round(porcentaje*100) / 100 // Redondear a 2 decimales
	} else {
		porcentaje = 100.0 // Si la exp. próxima es 0, asumimos nivel máximo
	}

	return models.ProgressResponse{
		NivelActual:          baseData.NivelActual,
		ExperienciaActual:    baseData.ExperienciaActual,
		ExperienciaProximo:   baseData.ExperienciaProxima,
		PorcentajeAvance:     porcentaje,
		RecompensasObtenidas: unlocked,
		ProximasRecompensas:  nextRewards,
	}, nil
}