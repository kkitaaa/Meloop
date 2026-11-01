package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ExperienceRepository interface {
	AddExperienceAtomic(ctx context.Context, userID string, xpToAdd int) (int, error)
	AddExperienceWithRetry(ctx context.Context, userID string, xpToAdd int, maxRetries int) (int, error)
}

type PostgresExperienceRepository struct {
	db *sql.DB
}

func NewPostgresExperienceRepository(db *sql.DB) *PostgresExperienceRepository {
	return &PostgresExperienceRepository{db: db}
}

// AddExperienceAtomic actualiza la experiencia usando los nombres reales del MER (tabla usuario)
func (r *PostgresExperienceRepository) AddExperienceAtomic(ctx context.Context, userID string, xpToAdd int) (int, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, fmt.Errorf("error al iniciar transacción: %w", err)
	}
	defer tx.Rollback()

	// Consulta adaptada estrictamente al diagrama MER (tabla usuario, id_usuario)
	query := `
		UPDATE usuario 
		SET experiencia = experiencia + $1 
		WHERE id_usuario = $2 
		RETURNING experiencia;
	`

	var newXP int
	err = tx.QueryRowContext(ctx, query, xpToAdd, userID).Scan(&newXP)
	if err != nil {
		return 0, fmt.Errorf("error al actualizar experiencia: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("error al confirmar transacción: %w", err)
	}

	return newXP, nil
}

// AddExperienceWithRetry agrega reintentos en caso de fallos de red
func (r *PostgresExperienceRepository) AddExperienceWithRetry(ctx context.Context, userID string, xpToAdd int, maxRetries int) (int, error) {
	var newXP int
	var err error

	backoff := 100 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		newXP, err = r.AddExperienceAtomic(ctx, userID, xpToAdd)
		if err == nil {
			return newXP, nil
		}

		if ctx.Err() != nil {
			return 0, ctx.Err()
		}

		time.Sleep(backoff)
		backoff *= 2
	}

	return 0, fmt.Errorf("se superó el máximo de reintentos (%d): %w", maxRetries, err)
}