package repositories

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/notification-service/models"
)

func TestNotificationPersistenceAndReadState(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL no está configurada; se omite la prueba de integración")
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(poolConfig.ConnConfig.Database), "test") {
		t.Fatal("TEST_DATABASE_URL debe apuntar a una base de datos cuyo nombre incluya 'test'")
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()
	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := "notification-test-" + runID
	_, err = pool.Exec(ctx, `
		INSERT INTO usuario (id_usuario, username, correo, contrasena_hash)
		VALUES ($1, $2, $3, $4)
	`, userID, "notificationtest"+runID, userID+"@example.test", "test-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := pool.Exec(context.Background(), `DELETE FROM usuario WHERE id_usuario = $1`, userID); cleanupErr != nil {
			t.Errorf("cleanup notification test user: %v", cleanupErr)
		}
	})

	_, err = pool.Exec(ctx, `
		INSERT INTO configuracion_notificacion (id_configuracion, id_usuario, tipo_notificacion, habilitada)
		VALUES ($1, $2, 'LIKE', FALSE)
	`, "notification-config-"+runID, userID)
	if err != nil {
		t.Fatal(err)
	}

	repository := NewNotificationRepository(pool)
	disabled := models.Notification{
		ID:             "notification-disabled-" + runID,
		EventID:        "event-disabled-" + runID,
		Recipient:      userID,
		Type:           "post.liked",
		PreferenceType: "LIKE",
		Data:           []byte(`{"postId":"post-1"}`),
	}
	if err := repository.Create(ctx, disabled); err != nil {
		t.Fatal(err)
	}
	var notificationCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM notificacion WHERE id_usuario = $1`, userID).Scan(&notificationCount); err != nil {
		t.Fatal(err)
	}
	if notificationCount != 0 {
		t.Fatalf("disabled notification count = %d, want 0", notificationCount)
	}

	if _, err := pool.Exec(ctx, `
		UPDATE configuracion_notificacion SET habilitada = TRUE
		WHERE id_usuario = $1 AND tipo_notificacion = 'LIKE'
	`, userID); err != nil {
		t.Fatal(err)
	}
	enabled := disabled
	enabled.ID = "notification-enabled-" + runID
	enabled.EventID = "event-enabled-" + runID
	if err := repository.Create(ctx, enabled); err != nil {
		t.Fatal(err)
	}

	notifications, err := repository.ListByUserID(ctx, userID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(notifications) != 1 || notifications[0].Read {
		t.Fatalf("notifications = %+v, want one unread notification", notifications)
	}
	if err := repository.MarkRead(ctx, userID, notifications[0].ID); err != nil {
		t.Fatal(err)
	}
	updated, err := repository.ListByUserID(ctx, userID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !updated[0].Read {
		t.Fatal("notification should be marked as read")
	}
	defaultEnabled := models.Notification{
		ID:             "notification-default-" + runID,
		EventID:        "event-default-" + runID,
		Recipient:      userID,
		Type:           "comment.created",
		PreferenceType: "COMENTARIO",
		Data:           []byte(`{"postId":"post-2"}`),
	}
	if err := repository.Create(ctx, defaultEnabled); err != nil {
		t.Fatal(err)
	}
	if count, err := repository.MarkAllRead(ctx, userID); err != nil || count != 1 {
		t.Fatalf("MarkAllRead() = (%d, %v), want (1, nil)", count, err)
	}
	updated, err = repository.ListByUserID(ctx, userID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, notification := range updated {
		if !notification.Read {
			t.Errorf("notification %q should be marked as read", notification.ID)
		}
	}
	if err := repository.MarkRead(ctx, "another-user", notifications[0].ID); err != ErrNotificationNotFound {
		t.Fatalf("MarkRead for another user error = %v, want ErrNotificationNotFound", err)
	}
}
