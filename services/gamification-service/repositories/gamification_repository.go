package repositories

import (
	"github.com/meloop/gamification-service/models"
	"gorm.io/gorm"
)

type GamificationRepository struct {
	db *gorm.DB
}

func NewGamificationRepository(db *gorm.DB) *GamificationRepository {
	return &GamificationRepository{db: db}
}

// GetOrCreateProgress busca al usuario, si no existe le crea un progreso en Nivel 1 con 0 XP
func (r *GamificationRepository) GetOrCreateProgress(userID string) (models.UserProgress, error) {
	var progress models.UserProgress
	err := r.db.Where("user_id = ?", userID).FirstOrCreate(&progress, models.UserProgress{UserID: userID, TotalXP: 0, Level: 1}).Error
	return progress, err
}

func (r *GamificationRepository) UpdateProgress(progress models.UserProgress) error {
	return r.db.Save(&progress).Error
}

func (r *GamificationRepository) GetLevelRule(level int) (models.LevelRule, error) {
	var rule models.LevelRule
	err := r.db.Where("level = ?", level).First(&rule).Error
	return rule, err
}

func (r *GamificationRepository) UnlockReward(reward models.UserReward) error {
	return r.db.Create(&reward).Error
}