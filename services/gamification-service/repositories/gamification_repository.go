package repositories

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
	ListRewards(ctx context.Context) ([]models.Reward, error)
	CreateReward(ctx context.Context, reward *models.RewardCreateRequest) (*models.Reward, error)
	UpdateReward(ctx context.Context, id string, reward *models.RewardUpdateRequest) (*models.Reward, error)
	DeleteReward(ctx context.Context, id string) error
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

func (r *postgresGamificationRepository) ensureRewardColumns(ctx context.Context) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}
	queries := []string{
		`ALTER TABLE RECOMPENSA ADD COLUMN IF NOT EXISTS nombre VARCHAR(100)`,
		`ALTER TABLE RECOMPENSA ADD COLUMN IF NOT EXISTS descripcion TEXT`,
		`ALTER TABLE RECOMPENSA ADD COLUMN IF NOT EXISTS habilitada BOOLEAN NOT NULL DEFAULT TRUE`,
		`ALTER TABLE RECOMPENSA ADD COLUMN IF NOT EXISTS disponible BOOLEAN NOT NULL DEFAULT TRUE`,
	}
	for _, q := range queries {
		if _, err := r.pool.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (r *postgresGamificationRepository) ListRewards(ctx context.Context) ([]models.Reward, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	if err := r.ensureRewardColumns(ctx); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id_recompensa, tipo, COALESCE(nombre, tipo), COALESCE(descripcion, ''), COALESCE(id_nivel, 1),
			COALESCE(habilitada, TRUE), COALESCE(disponible, TRUE)
		FROM RECOMPENSA
		ORDER BY id_nivel ASC, tipo ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rewards := make([]models.Reward, 0)
	for rows.Next() {
		var reward models.Reward
		if err := rows.Scan(&reward.IDRecompensa, &reward.Tipo, &reward.Nombre, &reward.Descripcion, &reward.IDNivel, &reward.Habilitada, &reward.Disponible); err != nil {
			return nil, err
		}
		rewards = append(rewards, reward)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rewards, nil
}

func (r *postgresGamificationRepository) CreateReward(ctx context.Context, reward *models.RewardCreateRequest) (*models.Reward, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	if reward == nil {
		return nil, errors.New("reward payload is required")
	}
	if err := r.ensureRewardColumns(ctx); err != nil {
		return nil, err
	}

	id := generateID("reward")
	var created models.Reward
	query := `
		INSERT INTO RECOMPENSA (id_recompensa, tipo, nombre, descripcion, id_nivel, habilitada, disponible)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id_recompensa, tipo, COALESCE(nombre, tipo), COALESCE(descripcion, ''), id_nivel,
			COALESCE(habilitada, TRUE), COALESCE(disponible, TRUE)
	`
	err := r.pool.QueryRow(ctx, query, id, reward.Tipo, reward.Nombre, reward.Descripcion, reward.IDNivel, reward.Habilitada, reward.Disponible).Scan(
		&created.IDRecompensa,
		&created.Tipo,
		&created.Nombre,
		&created.Descripcion,
		&created.IDNivel,
		&created.Habilitada,
		&created.Disponible,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *postgresGamificationRepository) UpdateReward(ctx context.Context, id string, reward *models.RewardUpdateRequest) (*models.Reward, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	if reward == nil {
		return nil, errors.New("reward payload is required")
	}
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("reward id is required")
	}
	if err := r.ensureRewardColumns(ctx); err != nil {
		return nil, err
	}

	updates := make([]string, 0, 6)
	values := make([]interface{}, 0, 7)
	index := 1

	if reward.Tipo != "" {
		updates = append(updates, "tipo = $"+strconv.Itoa(index))
		values = append(values, reward.Tipo)
		index++
	}
	if reward.Nombre != "" {
		updates = append(updates, "nombre = $"+strconv.Itoa(index))
		values = append(values, reward.Nombre)
		index++
	}
	if reward.Descripcion != "" {
		updates = append(updates, "descripcion = $"+strconv.Itoa(index))
		values = append(values, reward.Descripcion)
		index++
	}
	if reward.IDNivel != 0 {
		updates = append(updates, "id_nivel = $"+strconv.Itoa(index))
		values = append(values, reward.IDNivel)
		index++
	}
	if reward.Habilitada != nil {
		updates = append(updates, "habilitada = $"+strconv.Itoa(index))
		values = append(values, *reward.Habilitada)
		index++
	}
	if reward.Disponible != nil {
		updates = append(updates, "disponible = $"+strconv.Itoa(index))
		values = append(values, *reward.Disponible)
		index++
	}
	if len(updates) == 0 {
		return nil, nil
	}

	query := "UPDATE RECOMPENSA SET " + strings.Join(updates, ", ") + " WHERE id_recompensa = $" + strconv.Itoa(index) + " RETURNING id_recompensa, tipo, COALESCE(nombre, tipo), COALESCE(descripcion, ''), id_nivel, COALESCE(habilitada, TRUE), COALESCE(disponible, TRUE)"
	values = append(values, id)

	var created models.Reward
	err := r.pool.QueryRow(ctx, query, values...).Scan(
		&created.IDRecompensa,
		&created.Tipo,
		&created.Nombre,
		&created.Descripcion,
		&created.IDNivel,
		&created.Habilitada,
		&created.Disponible,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &created, nil
}

func (r *postgresGamificationRepository) DeleteReward(ctx context.Context, id string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("reward id is required")
	}
	_, err := r.pool.Exec(ctx, `DELETE FROM RECOMPENSA WHERE id_recompensa = $1`, id)
	return err
}

func generateID(prefix string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
