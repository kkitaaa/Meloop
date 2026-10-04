package repositories

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/gamification-service/models"
)

var ErrRequiredLevelNotFound = errors.New("required level not found")

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) IsAdmin(ctx context.Context, userID string) (bool, error) {
	var isAdmin bool
	err := r.pool.QueryRow(ctx, `SELECT rol = 'ADMIN' FROM usuario WHERE id_usuario = $1`, userID).Scan(&isAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return isAdmin, err
}

func (r *AdminRepository) List(ctx context.Context, includeUnavailable bool) ([]models.Reward, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id_recompensa, nombre, descripcion, tipo, id_nivel, disponible
		FROM recompensa
		WHERE $1 OR disponible = TRUE
		ORDER BY id_nivel, nombre, id_recompensa`, includeUnavailable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rewards := make([]models.Reward, 0)
	for rows.Next() {
		var reward models.Reward
		if err := rows.Scan(&reward.ID, &reward.Name, &reward.Description, &reward.Type, &reward.RequiredLevel, &reward.Available); err != nil {
			return nil, err
		}
		rewards = append(rewards, reward)
	}
	return rewards, rows.Err()
}

func (r *AdminRepository) Get(ctx context.Context, rewardID string) (*models.Reward, error) {
	var reward models.Reward
	err := r.pool.QueryRow(ctx, `
		SELECT id_recompensa, nombre, descripcion, tipo, id_nivel, disponible
		FROM recompensa WHERE id_recompensa = $1`, rewardID).Scan(
		&reward.ID, &reward.Name, &reward.Description, &reward.Type, &reward.RequiredLevel, &reward.Available,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &reward, err
}

func (r *AdminRepository) Create(ctx context.Context, reward models.Reward) (*models.Reward, error) {
	id, err := newRewardID()
	if err != nil {
		return nil, err
	}
	var created models.Reward
	err = r.pool.QueryRow(ctx, `
		INSERT INTO recompensa(id_recompensa, nombre, descripcion, tipo, id_nivel, disponible)
		SELECT $1, $2, $3, $4, id_nivel, $6
		FROM nivel WHERE id_nivel = $5
		RETURNING id_recompensa, nombre, descripcion, tipo, id_nivel, disponible`,
		id, reward.Name, reward.Description, reward.Type, reward.RequiredLevel, reward.Available).Scan(
		&created.ID, &created.Name, &created.Description, &created.Type, &created.RequiredLevel, &created.Available,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRequiredLevelNotFound
	}
	return &created, err
}

func (r *AdminRepository) Update(ctx context.Context, rewardID string, reward models.Reward) (*models.Reward, error) {
	var levelExists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM nivel WHERE id_nivel = $1)`, reward.RequiredLevel).Scan(&levelExists); err != nil {
		return nil, err
	}
	if !levelExists {
		return nil, ErrRequiredLevelNotFound
	}
	var updated models.Reward
	err := r.pool.QueryRow(ctx, `
		UPDATE recompensa SET nombre = $2, descripcion = $3, tipo = $4, id_nivel = $5, disponible = $6
		WHERE id_recompensa = $1
		RETURNING id_recompensa, nombre, descripcion, tipo, id_nivel, disponible`,
		rewardID, reward.Name, reward.Description, reward.Type, reward.RequiredLevel, reward.Available).Scan(
		&updated.ID, &updated.Name, &updated.Description, &updated.Type, &updated.RequiredLevel, &updated.Available,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &updated, err
}

func (r *AdminRepository) SetAvailable(ctx context.Context, rewardID string, available bool) (*models.Reward, error) {
	var updated models.Reward
	err := r.pool.QueryRow(ctx, `
		UPDATE recompensa SET disponible = $2
		WHERE id_recompensa = $1
		RETURNING id_recompensa, nombre, descripcion, tipo, id_nivel, disponible`, rewardID, available).Scan(
		&updated.ID, &updated.Name, &updated.Description, &updated.Type, &updated.RequiredLevel, &updated.Available,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &updated, err
}

func newRewardID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate reward ID: %w", err)
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	encoded := fmt.Sprintf("%x", value)
	return strings.Join([]string{encoded[:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:]}, "-"), nil
}
