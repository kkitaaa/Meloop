package repositories

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"

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
