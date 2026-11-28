package repositories

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/notification-service/models"
)

var ErrNotificationNotFound = errors.New("notification not found")

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) ResolvePostOwner(ctx context.Context, postID string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `SELECT id_usuario FROM publicacion WHERE id_publicacion = $1`, postID).Scan(&userID)
	return userID, err
}

func (r *NotificationRepository) ResolveInteractionOwner(ctx context.Context, interactionID string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `SELECT id_usuario FROM comentario WHERE id_comentario = $1`, interactionID).Scan(&userID)
	return userID, err
}

func (r *NotificationRepository) Create(ctx context.Context, notification models.Notification) error {
	if r.pool == nil {
		return errors.New("database pool is not initialized")
	}
	if notification.ID == "" {
		id, err := newID()
		if err != nil {
			return fmt.Errorf("generate notification ID: %w", err)
		}
		notification.ID = id
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notificacion
			(id_notificacion, id_usuario, id_emisor, tipo, id_objetivo, leida, contador, id_evento, datos)
		SELECT $1, $2, $3, $4, $5, FALSE, 1, $6, $7::jsonb
		WHERE NOT EXISTS (
			SELECT 1
			FROM configuracion_notificacion
			WHERE id_usuario = $2
				AND UPPER(TRIM(tipo_notificacion)) = $8
				AND habilitada = FALSE
		)
		ON CONFLICT (id_evento) DO NOTHING`,
		notification.ID,
		notification.Recipient,
		notification.Actor,
		notification.Type,
		notification.TargetID,
		notification.EventID,
		notification.Data,
		notification.PreferenceType,
	)
	return err
}

func (r *NotificationRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]models.NotificationResponse, error) {
	if r.pool == nil {
		return nil, errors.New("database pool is not initialized")
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id_notificacion, id_usuario, id_emisor, tipo, id_objetivo, leida,
			contador, id_evento, datos, fecha_creacion
		FROM notificacion
		WHERE id_usuario = $1
		ORDER BY fecha_creacion DESC, id_notificacion DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]models.NotificationResponse, 0)
	for rows.Next() {
		var notification models.NotificationResponse
		if err := rows.Scan(
			&notification.ID,
			&notification.Recipient,
			&notification.Actor,
			&notification.Type,
			&notification.TargetID,
			&notification.Read,
			&notification.Count,
			&notification.EventID,
			&notification.Data,
			&notification.CreatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, notificationID string) error {
	if r.pool == nil {
		return errors.New("database pool is not initialized")
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE notificacion
		SET leida = TRUE
		WHERE id_notificacion = $1 AND id_usuario = $2
	`, notificationID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	if r.pool == nil {
		return 0, errors.New("database pool is not initialized")
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE notificacion
		SET leida = TRUE
		WHERE id_usuario = $1 AND leida = FALSE
	`, userID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func newID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	encoded := fmt.Sprintf("%x", value)
	return strings.Join([]string{encoded[:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:]}, "-"), nil
}
