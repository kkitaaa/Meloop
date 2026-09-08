package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/models"
)

// NotificationConfigRepository define las operaciones de persistencia para CONFIGURACION_NOTIFICACIONES
type NotificationConfigRepository interface {
	GetByUserID(ctx context.Context, userID string) ([]models.ConfiguracionNotificacion, error)
	CreateDefaults(ctx context.Context, userID string) ([]models.ConfiguracionNotificacion, error)
	Upsert(ctx context.Context, userID string, tipo string, habilitada bool) (*models.ConfiguracionNotificacion, error)
	UpsertBatch(ctx context.Context, userID string, items []models.NotificationSettingItem) ([]models.ConfiguracionNotificacion, error)
}

type postgresNotificationConfigRepository struct {
	pool *pgxpool.Pool
}

// NewNotificationConfigRepository crea una nueva instancia de NotificationConfigRepository
func NewNotificationConfigRepository(pool *pgxpool.Pool) NotificationConfigRepository {
	return &postgresNotificationConfigRepository{pool: pool}
}

// GetByUserID obtiene todas las configuraciones de notificación de un usuario.
// Si no existen configuraciones o faltan tipos, se inicializan con valores por defecto (habilitada = true).
func (r *postgresNotificationConfigRepository) GetByUserID(ctx context.Context, userID string) ([]models.ConfiguracionNotificacion, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	query := `
		SELECT id_configuracion, id_usuario::text, tipo_notificacion, habilitada
		FROM CONFIGURACION_NOTIFICACIONES
		WHERE id_usuario = $1
		ORDER BY id_configuracion ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ConfiguracionNotificacion
	foundTypes := make(map[string]bool)

	for rows.Next() {
		var item models.ConfiguracionNotificacion
		if err := rows.Scan(&item.IDConfiguracion, &item.IDUsuario, &item.TipoNotificacion, &item.Habilitada); err != nil {
			return nil, err
		}
		result = append(result, item)
		foundTypes[item.TipoNotificacion] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Si no hay configuraciones o falta alguno de los 9 tipos, inicializar los faltantes
	if len(result) < len(models.AllNotificationTypes) {
		return r.CreateDefaults(ctx, userID)
	}

	return result, nil
}

// CreateDefaults inserta los 9 tipos de notificación por defecto habilitados (true) si no existen
func (r *postgresNotificationConfigRepository) CreateDefaults(ctx context.Context, userID string) ([]models.ConfiguracionNotificacion, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	query := `
		INSERT INTO CONFIGURACION_NOTIFICACIONES (id_usuario, tipo_notificacion, habilitada)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (id_usuario, tipo_notificacion) DO NOTHING
	`

	for _, tipo := range models.AllNotificationTypes {
		_, err := r.pool.Exec(ctx, query, userID, tipo)
		if err != nil {
			return nil, err
		}
	}

	// Retornar la lista completa ya garantizada
	fetchQuery := `
		SELECT id_configuracion, id_usuario::text, tipo_notificacion, habilitada
		FROM CONFIGURACION_NOTIFICACIONES
		WHERE id_usuario = $1
		ORDER BY id_configuracion ASC
	`
	rows, err := r.pool.Query(ctx, fetchQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ConfiguracionNotificacion
	for rows.Next() {
		var item models.ConfiguracionNotificacion
		if err := rows.Scan(&item.IDConfiguracion, &item.IDUsuario, &item.TipoNotificacion, &item.Habilitada); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

// Upsert inserta o actualiza el estado de un tipo de notificación para un usuario
func (r *postgresNotificationConfigRepository) Upsert(ctx context.Context, userID string, tipo string, habilitada bool) (*models.ConfiguracionNotificacion, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	var item models.ConfiguracionNotificacion
	query := `
		INSERT INTO CONFIGURACION_NOTIFICACIONES (id_usuario, tipo_notificacion, habilitada)
		VALUES ($1, $2, $3)
		ON CONFLICT (id_usuario, tipo_notificacion)
		DO UPDATE SET habilitada = EXCLUDED.habilitada
		RETURNING id_configuracion, id_usuario::text, tipo_notificacion, habilitada
	`
	err := r.pool.QueryRow(ctx, query, userID, tipo, habilitada).Scan(
		&item.IDConfiguracion,
		&item.IDUsuario,
		&item.TipoNotificacion,
		&item.Habilitada,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UpsertBatch inserta o actualiza múltiples tipos de notificación en una única operación
func (r *postgresNotificationConfigRepository) UpsertBatch(ctx context.Context, userID string, items []models.NotificationSettingItem) ([]models.ConfiguracionNotificacion, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	upsertQuery := `
		INSERT INTO CONFIGURACION_NOTIFICACIONES (id_usuario, tipo_notificacion, habilitada)
		VALUES ($1, $2, $3)
		ON CONFLICT (id_usuario, tipo_notificacion)
		DO UPDATE SET habilitada = EXCLUDED.habilitada
	`

	for _, item := range items {
		if item.Habilitada != nil {
			_, err := tx.Exec(ctx, upsertQuery, userID, item.TipoNotificacion, *item.Habilitada)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByUserID(ctx, userID)
}
