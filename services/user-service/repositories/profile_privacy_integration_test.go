package repositories

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestProfileRepositoryPrivacyVisibilityPostgres(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL no está configurada; se omite la prueba de integración")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(config.ConnConfig.Database), "test",
		"TEST_DATABASE_URL debe apuntar a una base de datos cuyo nombre incluya 'test'")

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(context.Background()))

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	ownerID := "privacy-owner-" + runID
	friendID := "privacy-friend-" + runID
	strangerID := "privacy-stranger-" + runID
	userIDs := []string{ownerID, friendID, strangerID}
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(), `DELETE FROM usuario WHERE id_usuario = ANY($1)`, userIDs)
		pool.Close()
		require.NoError(t, cleanupErr)
	})

	for i, userID := range userIDs {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO usuario (id_usuario, username, correo, contrasena_hash)
			VALUES ($1, $2, $3, $4)`,
			userID,
			fmt.Sprintf("privacy%s%d", runID, i),
			fmt.Sprintf("privacy%s%d@example.test", runID, i),
			"test-password-hash",
		)
		require.NoError(t, err)
	}
	_, err = pool.Exec(context.Background(), `
		INSERT INTO amistad (id_amistad, id_usuario_1, id_usuario_2, estado)
		VALUES ($1, $2, $3, 'ACEPTADA')`,
		"privacy-friendship-"+runID, ownerID, friendID,
	)
	require.NoError(t, err)

	repo := NewProfileRepository(pool)
	allowed, err := repo.CanViewVisibility(context.Background(), ownerID, friendID, "AMIGOS")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = repo.CanViewVisibility(context.Background(), ownerID, strangerID, "AMIGOS")
	require.NoError(t, err)
	require.False(t, allowed)

	allowed, err = repo.CanViewVisibility(context.Background(), ownerID, strangerID, "PUBLICO")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = repo.CanViewVisibility(context.Background(), ownerID, strangerID, "PRIVADO")
	require.NoError(t, err)
	require.False(t, allowed)

	allowed, err = repo.CanViewVisibility(context.Background(), ownerID, ownerID, "PRIVADO")
	require.NoError(t, err)
	require.True(t, allowed)
}
