package repositories

import (
	"errors"

	"github.com/meloop/gamification-service/models"
	"gorm.io/gorm"
)

type GamificationRepository struct {
	db *gorm.DB
}

func NewGamificationRepository(db *gorm.DB) *GamificationRepository {
	return &GamificationRepository{db: db}
}

// GetOrCreateProgress busca el progreso por UUID. Si no existe, lo crea con Nivel 1 y 0 XP.
func (r *GamificationRepository) GetOrCreateProgress(userID string) (models.UserProgress, error) {
	var progress models.UserProgress

	// 1. Intentar buscar si el usuario ya existe
	err := r.db.Where("user_id = ?", userID).First(&progress).Error
	if err == nil {
		return progress, nil
	}

	// 2. Si el registro no existe, crearlo
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newProgress := models.UserProgress{
			UserID:  userID,
			TotalXP: 0,
			Level:   1,
		}
		if createErr := r.db.Create(&newProgress).Error; createErr != nil {
			return models.UserProgress{}, createErr
		}
		return newProgress, nil
	}

	// 3. Retornar cualquier otro error de BD
	return models.UserProgress{}, err
}

// UpdateProgress actualiza los datos del usuario usando Save
func (r *GamificationRepository) UpdateProgress(progress models.UserProgress) error {
	return r.db.Save(&progress).Error
}

// GetLevelRules obtiene todas las reglas de nivel ordenadas
func (r *GamificationRepository) GetLevelRules() ([]models.LevelRule, error) {
	var rules []models.LevelRule
	err := r.db.Order("level asc").Find(&rules).Error
	return rules, err
}

// GetLevelRule obtiene una regla de nivel específica
func (r *GamificationRepository) GetLevelRule(level int) (models.LevelRule, error) {
	var rule models.LevelRule
	err := r.db.Where("level = ?", level).First(&rule).Error
	return rule, err
}

// UnlockReward registra una recompensa entregada
func (r *GamificationRepository) UnlockReward(reward models.UserReward) error {
	return r.db.Create(&reward).Error
}

// GetUserRewards lista las recompensas desbloqueadas por el usuario
func (r *GamificationRepository) GetUserRewards(userID string) ([]models.UserReward, error) {
	var rewards []models.UserReward
	err := r.db.Where("user_id = ?", userID).Find(&rewards).Error
	return rewards, err
}