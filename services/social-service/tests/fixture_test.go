package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/social-service/repositories"
	"github.com/meloop/social-service/services"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	pool       *pgxpool.Pool
	userIDs    [4]string
	friendship int64
}

func setupFixture(t *testing.T) *fixture {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL no está configurada; se omite la prueba de integración")
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(poolConfig.ConnConfig.Database), "test",
		"TEST_DATABASE_URL debe apuntar a una base de datos cuyo nombre incluya 'test'")

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(context.Background()))

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	testFixture := &fixture{pool: pool}
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(),
			`DELETE FROM usuario WHERE id_usuario = ANY($1)`, testFixture.userIDs[:])
		pool.Close()
		require.NoError(t, cleanupErr)
	})

	for i := range testFixture.userIDs {
		testFixture.userIDs[i] = fmt.Sprintf("social-test-%s-%d", runID, i)
		_, err := pool.Exec(context.Background(), `
			INSERT INTO usuario (id_usuario, username, correo, contrasena_hash)
			VALUES ($1, $2, $3, $4)`,
			testFixture.userIDs[i],
			fmt.Sprintf("socialtest%s%d", runID, i),
			fmt.Sprintf("socialtest%s%d@example.test", runID, i),
			"test-password-hash",
		)
		require.NoError(t, err)
	}
	return testFixture
}

func (f *fixture) databaseService() services.FriendshipService {
	repository := repositories.NewFriendshipRepository(f.pool)
	return services.NewFriendshipService(repository, nil)
}

func (f *fixture) createFriendship(t *testing.T, userA, userB int, status string) {
	t.Helper()
	f.friendship++
	id := time.Now().UnixNano()%2_000_000_000 + f.friendship
	_, err := f.pool.Exec(context.Background(), `
		INSERT INTO amistad (id_amistad, id_usuario_1, id_usuario_2, estado)
		VALUES ($1, $2, $3, $4)`,
		fmt.Sprintf("%d", id), f.userIDs[userA], f.userIDs[userB], status)
	require.NoError(t, err)
}
