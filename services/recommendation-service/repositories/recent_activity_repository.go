package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecentActivity struct {
	ID     string
	UserID string
	PostID string
	Type   string
}

type RecentActivityRepository interface {
	RecordRecentActivity(ctx context.Context, activity RecentActivity) error
}

type postgresRecentActivityRepository struct {
	pool *pgxpool.Pool
}

func NewRecentActivityRepository(pool *pgxpool.Pool) RecentActivityRepository {
	return &postgresRecentActivityRepository{pool: pool}
}

func (repository *postgresRecentActivityRepository) RecordRecentActivity(ctx context.Context, activity RecentActivity) error {
	if repository.pool == nil {
		return errors.New("database connection pool is not initialized")
	}
	_, err := repository.pool.Exec(ctx, `
		INSERT INTO INTERACCION (id_interaccion, id_usuario, id_publicacion, tipo)
		SELECT $1, $2, $3, $4
		WHERE NOT EXISTS (
			SELECT 1
			FROM CONFIGURACION_PRIVACIDAD privacy
			WHERE privacy.id_usuario = $2
			  AND UPPER(TRIM(COALESCE(privacy.visibilidad_interacciones, 'PUBLICO'))) = 'PRIVADO'
		)
		ON CONFLICT (id_interaccion) DO NOTHING`,
		activity.ID, activity.UserID, activity.PostID, activity.Type)
	if err != nil {
		return fmt.Errorf("insert recent user activity: %w", err)
	}
	return nil
}
