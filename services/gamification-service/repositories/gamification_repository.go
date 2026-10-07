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
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/gamification-service/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidExperience = errors.New("invalid experience amount")
)

type GamificationRepository interface {
	AddExperience(ctx context.Context, userID string, amount int) (*models.ExperienceResult, error)
	GetUserProgress(ctx context.Context, userID string) (*models.UserProgress, error)
}

type postgresGamificationRepository struct {
	pool *pgxpool.Pool
}

func NewGamificationRepository(pool *pgxpool.Pool) GamificationRepository {
	return &postgresGamificationRepository{pool: pool}
}

func (r *postgresGamificationRepository) AddExperience(ctx context.Context, userID string, amount int) (*models.ExperienceResult, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	if amount <= 0 {
		return nil, ErrInvalidExperience
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentXP int
	var currentNivelID int
	var currentLevelNumber int

	err = tx.QueryRow(ctx, `
		SELECT u.experiencia, COALESCE(u.id_nivel, 1), COALESCE(n.numero, 1)
		FROM usuario u
		LEFT JOIN nivel n ON u.id_nivel = n.id_nivel
		WHERE u.id_usuario = $1
		FOR UPDATE OF u`, userID).Scan(&currentXP, &currentNivelID, &currentLevelNumber)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user xp: %w", err)
	}

	newXP := currentXP + amount
	var newNivelID int
	var newLevelNumber int
	var reqXP int

	err = tx.QueryRow(ctx, `
		SELECT id_nivel, numero, experiencia_requerida
		FROM nivel
		WHERE experiencia_requerida <= $1
		ORDER BY numero DESC
		LIMIT 1`, newXP).Scan(&newNivelID, &newLevelNumber, &reqXP)

	if errors.Is(err, pgx.ErrNoRows) {
		newNivelID = currentNivelID
		newLevelNumber = currentLevelNumber
	} else if err != nil {
		return nil, fmt.Errorf("calculate level: %w", err)
	}

	leveledUp := newLevelNumber > currentLevelNumber

	_, err = tx.Exec(ctx, `
		UPDATE usuario
		SET experiencia = $1, id_nivel = $2
		WHERE id_usuario = $3`, newXP, newNivelID, userID)
	if err != nil {
		return nil, fmt.Errorf("update user xp and level: %w", err)
	}

	unlockedRewards := make([]models.Reward, 0)
	if leveledUp {
		rows, err := tx.Query(ctx, `
			SELECT id_recompensa, tipo, id_nivel
			FROM recompensa
			WHERE id_nivel > $1 AND id_nivel <= $2
			ORDER BY id_nivel ASC, id_recompensa ASC`, currentNivelID, newNivelID)
		if err != nil {
			return nil, fmt.Errorf("query rewards: %w", err)
		}
		defer rows.Close()

		var rewards []models.Reward
		for rows.Next() {
			var rew models.Reward
			if err := rows.Scan(&rew.IDRecompensa, &rew.Tipo, &rew.IDNivel); err != nil {
				return nil, fmt.Errorf("scan reward: %w", err)
			}
			rewards = append(rewards, rew)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rewards iteration: %w", err)
		}

		for _, rew := range rewards {
			var alreadyInInventory bool
			err := tx.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM inventario
					WHERE id_usuario = $1 AND id_recompensa = $2
				)`, userID, rew.IDRecompensa).Scan(&alreadyInInventory)
			if err != nil {
				return nil, fmt.Errorf("check inventory: %w", err)
			}

			if !alreadyInInventory {
				invID := generateID("inv")
				_, err = tx.Exec(ctx, `
					INSERT INTO inventario (id_inventario, id_usuario, id_recompensa, equipada)
					VALUES ($1, $2, $3, false)`, invID, userID, rew.IDRecompensa)
				if err != nil {
					return nil, fmt.Errorf("insert inventory item: %w", err)
				}
				unlockedRewards = append(unlockedRewards, rew)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &models.ExperienceResult{
		UserID:          userID,
		PreviousXP:      currentXP,
		CurrentXP:       newXP,
		PreviousLevel:   currentLevelNumber,
		CurrentLevel:    newLevelNumber,
		LeveledUp:       leveledUp,
		UnlockedRewards: unlockedRewards,
	}, nil
}

func (r *postgresGamificationRepository) GetUserProgress(ctx context.Context, userID string) (*models.UserProgress, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	var progress models.UserProgress
	progress.UserID = userID
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(u.id_nivel, 1), COALESCE(n.numero, 1), u.experiencia
		FROM usuario u
		LEFT JOIN nivel n ON u.id_nivel = n.id_nivel
		WHERE u.id_usuario = $1`, userID).Scan(&progress.IDNivel, &progress.LevelNumber, &progress.Experiencia)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user progress: %w", err)
	}

	return &progress, nil
}

func generateID(prefix string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
