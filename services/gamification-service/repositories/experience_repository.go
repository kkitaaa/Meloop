package repositories

import (
	"context"
	"database/sql"
	"fmt"
)

type ExperienceRepository struct {
	db *sql.DB
}

// Inyectamos la conexión real a la base de datos
func NewExperienceRepository(db *sql.DB) *ExperienceRepository {
	return &ExperienceRepository{db: db}
}

func (r *ExperienceRepository) AddExperience(ctx context.Context, userID string, xpToAdd int) error {
	// Asumimos que la tabla se llama 'users' y la columna 'xp'
	query := `UPDATE usuario SET experiencia = experiencia + $1 WHERE id_usuario = $2`
	
	result, err := r.db.ExecContext(ctx, query, xpToAdd, userID)
	if err != nil {
		return fmt.Errorf("error ejecutando el query de XP: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("usuario no encontrado en la base de datos: %s", userID)
	}

	return nil
}