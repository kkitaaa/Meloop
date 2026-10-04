package repositories

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/notification-service/models"
)

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
	err := r.pool.QueryRow(ctx, `SELECT id_usuario FROM interaccion WHERE id_interaccion = $1`, interactionID).Scan(&userID)
	return userID, err
}

func (r *NotificationRepository) IsEnabled(ctx context.Context, userID, notificationType string) (bool, error) {
	var enabled bool
	err := r.pool.QueryRow(ctx, `
		SELECT habilitada
		FROM configuracion_notificacion
		WHERE id_usuario = $1 AND tipo_notificacion = $2`, userID, notificationType).Scan(&enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}
	return enabled, err
}

func (r *NotificationRepository) List(ctx context.Context, userID string, limit, offset int) ([]models.NotificationRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id_notificacion, id_emisor, tipo, id_objetivo, leida, contador, fecha_creacion, datos
		FROM notificacion
		WHERE id_usuario = $1
		ORDER BY fecha_creacion DESC, id_notificacion DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]models.NotificationRecord, 0)
	for rows.Next() {
		var notification models.NotificationRecord
		var senderID, targetID *string
		var createdAt time.Time
		if err := rows.Scan(
			&notification.ID,
			&senderID,
			&notification.Type,
			&targetID,
			&notification.Read,
			&notification.Count,
			&createdAt,
			&notification.Data,
		); err != nil {
			return nil, err
		}
		if senderID != nil {
			notification.SenderID = *senderID
		}
		if targetID != nil {
			notification.TargetID = *targetID
		}
		notification.CreatedAt = createdAt.UTC().Format(time.RFC3339Nano)
		notifications = append(notifications, notification)
	}
	return notifications, rows.Err()
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, notificationID string) (bool, error) {
	result, err := r.pool.Exec(ctx, `
		UPDATE notificacion SET leida = TRUE
		WHERE id_notificacion = $1 AND id_usuario = $2`, notificationID, userID)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	result, err := r.pool.Exec(ctx, `
		UPDATE notificacion SET leida = TRUE
		WHERE id_usuario = $1 AND leida = FALSE`, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *NotificationRepository) ListPreferences(ctx context.Context, userID string) ([]models.NotificationPreference, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT tipo_notificacion, habilitada
		FROM configuracion_notificacion
		WHERE id_usuario = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	enabledByType := make(map[string]bool)
	for rows.Next() {
		var notificationType string
		var enabled bool
		if err := rows.Scan(&notificationType, &enabled); err != nil {
			return nil, err
		}
		enabledByType[notificationType] = enabled
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	preferences := make([]models.NotificationPreference, 0, len(models.SupportedNotificationTypes()))
	for _, notificationType := range models.SupportedNotificationTypes() {
		enabled, exists := enabledByType[notificationType]
		if !exists {
			enabled = true
		}
		preferences = append(preferences, models.NotificationPreference{Type: notificationType, Enabled: enabled})
	}
	return preferences, nil
}

func (r *NotificationRepository) SetPreference(ctx context.Context, userID, notificationType string, enabled bool) error {
	id, err := newID()
	if err != nil {
		return fmt.Errorf("generate preference ID: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO configuracion_notificacion
			(id_configuracion, id_usuario, tipo_notificacion, habilitada)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id_usuario, tipo_notificacion)
		DO UPDATE SET habilitada = EXCLUDED.habilitada`, id, userID, notificationType, enabled)
	return err
}

func (r *NotificationRepository) Create(ctx context.Context, notification models.Notification) error {
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
		VALUES ($1, $2, $3, $4, $5, FALSE, 1, $6, $7::jsonb)
		ON CONFLICT (id_evento) DO NOTHING`,
		notification.ID,
		notification.Recipient,
		notification.Actor,
		notification.Type,
		notification.TargetID,
		notification.EventID,
		notification.Data,
	)
	return err
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
