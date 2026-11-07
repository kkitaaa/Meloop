package repositories

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/models"
)

// NotificationConfigRepository define las operaciones de persistencia para CONFIGURACION_NOTIFICACION
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

func (r *postgresNotificationConfigRepository) GetByUserID(ctx context.Context, userID string) ([]models.ConfiguracionNotificacion, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	result, err := r.queryByUserID(ctx, r.pool, userID)
	if err != nil {
		return nil, err
	}
	if !hasAllNotificationTypes(result) {
		return r.CreateDefaults(ctx, userID)
	}
	return result, nil
}

func (r *postgresNotificationConfigRepository) CreateDefaults(ctx context.Context, userID string) ([]models.ConfiguracionNotificacion, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := lockNotificationSettings(ctx, tx, userID); err != nil {
		return nil, err
	}
	for _, tipo := range models.AllNotificationTypes {
		var exists bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM CONFIGURACION_NOTIFICACION
				WHERE id_usuario = $1 AND tipo_notificacion = $2
			)
		`, userID, tipo).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}

		id, err := newNotificationConfigID()
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO CONFIGURACION_NOTIFICACION
				(id_configuracion, id_usuario, tipo_notificacion, habilitada)
			VALUES ($1, $2, $3, TRUE)
		`, id, userID, tipo)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.queryByUserID(ctx, r.pool, userID)
}

func (r *postgresNotificationConfigRepository) Upsert(ctx context.Context, userID string, tipo string, habilitada bool) (*models.ConfiguracionNotificacion, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := lockNotificationSettings(ctx, tx, userID); err != nil {
		return nil, err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE CONFIGURACION_NOTIFICACION
		SET habilitada = $3
		WHERE id_usuario = $1 AND tipo_notificacion = $2
	`, userID, tipo, habilitada)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		id, err := newNotificationConfigID()
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO CONFIGURACION_NOTIFICACION
				(id_configuracion, id_usuario, tipo_notificacion, habilitada)
			VALUES ($1, $2, $3, $4)
		`, id, userID, tipo, habilitada)
		if err != nil {
			return nil, err
		}
	}

	item, err := queryNotificationSetting(ctx, tx, userID, tipo)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return item, nil
}

func (r *postgresNotificationConfigRepository) UpsertBatch(ctx context.Context, userID string, items []models.NotificationSettingItem) ([]models.ConfiguracionNotificacion, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := lockNotificationSettings(ctx, tx, userID); err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.Habilitada == nil {
			continue
		}
		tag, err := tx.Exec(ctx, `
			UPDATE CONFIGURACION_NOTIFICACION
			SET habilitada = $3
			WHERE id_usuario = $1 AND tipo_notificacion = $2
		`, userID, item.TipoNotificacion, *item.Habilitada)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() > 0 {
			continue
		}

		id, err := newNotificationConfigID()
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO CONFIGURACION_NOTIFICACION
				(id_configuracion, id_usuario, tipo_notificacion, habilitada)
			VALUES ($1, $2, $3, $4)
		`, id, userID, item.TipoNotificacion, *item.Habilitada)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByUserID(ctx, userID)
}

type notificationQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type notificationRowQueryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *postgresNotificationConfigRepository) queryByUserID(ctx context.Context, queryer notificationQueryer, userID string) ([]models.ConfiguracionNotificacion, error) {
	rows, err := queryer.Query(ctx, `
		SELECT id_configuracion, id_usuario::text, tipo_notificacion, habilitada
		FROM CONFIGURACION_NOTIFICACION
		WHERE id_usuario = $1
		ORDER BY id_configuracion ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.ConfiguracionNotificacion, 0, len(models.AllNotificationTypes))
	foundTypes := make(map[string]bool)
	for rows.Next() {
		var item models.ConfiguracionNotificacion
		if err := rows.Scan(&item.IDConfiguracion, &item.IDUsuario, &item.TipoNotificacion, &item.Habilitada); err != nil {
			return nil, err
		}
		if foundTypes[item.TipoNotificacion] {
			continue
		}
		foundTypes[item.TipoNotificacion] = true
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func queryNotificationSetting(ctx context.Context, queryer notificationRowQueryer, userID, tipo string) (*models.ConfiguracionNotificacion, error) {
	var item models.ConfiguracionNotificacion
	err := queryer.QueryRow(ctx, `
		SELECT id_configuracion, id_usuario::text, tipo_notificacion, habilitada
		FROM CONFIGURACION_NOTIFICACION
		WHERE id_usuario = $1 AND tipo_notificacion = $2
		ORDER BY id_configuracion ASC
		LIMIT 1
	`, userID, tipo).Scan(&item.IDConfiguracion, &item.IDUsuario, &item.TipoNotificacion, &item.Habilitada)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func lockNotificationSettings(ctx context.Context, tx pgx.Tx, userID string) error {
	var existingID string
	return tx.QueryRow(ctx, `
		SELECT id_usuario::text
		FROM USUARIO
		WHERE id_usuario = $1
		FOR UPDATE
	`, userID).Scan(&existingID)
}

func newNotificationConfigID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", fmt.Errorf("no se pudo generar el ID de configuración de notificación: %w", err)
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16]), nil
}

func hasAllNotificationTypes(settings []models.ConfiguracionNotificacion) bool {
	foundTypes := make(map[string]bool, len(settings))
	for _, setting := range settings {
		foundTypes[setting.TipoNotificacion] = true
	}
	for _, tipo := range models.AllNotificationTypes {
		if !foundTypes[tipo] {
			return false
		}
	}
	return true
}
