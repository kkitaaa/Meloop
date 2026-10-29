package repositories

import (
	"github.com/meloop/gamification-service/models"
	"gorm.io/gorm"
)

type ProgressRepository interface {
	GetBaseProgress(userID string) (models.UserProgressData, error)
	GetUnlockedRewards(userID string) ([]models.RewardInfo, error)
	GetNextRewards(currentLevel int) ([]models.RewardInfo, error)
}

type PostgresProgressRepository struct {
	db *gorm.DB
}

func NewPostgresProgressRepository(db *gorm.DB) *PostgresProgressRepository {
	return &PostgresProgressRepository{db: db}
}

func (r *PostgresProgressRepository) GetBaseProgress(userID string) (models.UserProgressData, error) {
	var data models.UserProgressData
	// Obtenemos nivel actual y la meta del siguiente nivel
	query := `
		SELECT 
			u.experiencia as experiencia_actual,
			n1.numero as nivel_actual,
			COALESCE(n2.experiencia_requerida, n1.experiencia_requerida) as experiencia_proxima
		FROM USUARIO u
		JOIN NIVEL n1 ON u.id_nivel = n1.id_nivel
		LEFT JOIN NIVEL n2 ON n2.numero = n1.numero + 1
		WHERE u.id_usuario = ?
	`
	err := r.db.Raw(query, userID).Scan(&data).Error
	return data, err
}

func (r *PostgresProgressRepository) GetUnlockedRewards(userID string) ([]models.RewardInfo, error) {
	var rewards []models.RewardInfo
	query := `
		SELECT r.id_recompensa as id, r.tipo
		FROM INVENTARIO i
		JOIN RECOMPENSA r ON i.id_recompensa = r.id_recompensa
		WHERE i.id_usuario = ?
	`
	err := r.db.Raw(query, userID).Scan(&rewards).Error
	return rewards, err
}

func (r *PostgresProgressRepository) GetNextRewards(currentLevel int) ([]models.RewardInfo, error) {
	var rewards []models.RewardInfo
	// Buscamos las recompensas vinculadas al siguiente nivel (currentLevel + 1)
	query := `
		SELECT r.id_recompensa as id, r.tipo
		FROM RECOMPENSA r
		JOIN NIVEL n ON r.id_nivel = n.id_nivel
		WHERE n.numero = ? + 1
	`
	err := r.db.Raw(query, currentLevel).Scan(&rewards).Error
	return rewards, err
}