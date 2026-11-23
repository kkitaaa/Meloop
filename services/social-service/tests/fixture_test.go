package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/social-service/client"
	"github.com/meloop/social-service/repositories"
	"github.com/meloop/social-service/services"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	pool       *pgxpool.Pool
	userIDs    [4]string
	extraUsers []string
	friendship int64
	runID      string
}

type testSessionValidator struct{}

func (testSessionValidator) ValidateToken(_ context.Context, token string) (string, error) {
	if token == "invalid" {
		return "", client.ErrInvalidSession
	}
	const prefix = "test-session:"
	if strings.HasPrefix(token, prefix) {
		val := strings.TrimPrefix(token, prefix)
		if val == "" {
			return "", client.ErrInvalidSession
		}
		return val, nil
	}
	return token, nil
}

func setTestSession(request *http.Request, userID string) {
	request.Header.Set("Authorization", "Bearer test-session:"+userID)
}

func setAuthenticatedUser(request *http.Request, userID string) {
	request.Header.Set("Authorization", "Bearer "+userID)
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

	var visibilityDefault, visibilityNullable string
	err = pool.QueryRow(context.Background(), `
		SELECT column_default, is_nullable
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = 'configuracion_privacidad'
		  AND column_name = 'visibilidad_amigos'`,
	).Scan(&visibilityDefault, &visibilityNullable)
	require.NoError(t, err, "visibilidad_amigos migration must be applied")
	require.Contains(t, visibilityDefault, "PUBLICO")
	require.Equal(t, "NO", visibilityNullable)

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	testFixture := &fixture{pool: pool, runID: runID}
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(),
			`DELETE FROM usuario WHERE id_usuario = ANY($1)`, append(testFixture.userIDs[:], testFixture.extraUsers...))
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

func (f *fixture) createUser(t *testing.T, username string) string {
	t.Helper()
	userID := fmt.Sprintf("social-test-%s-extra-%d", f.runID, len(f.extraUsers))
	f.extraUsers = append(f.extraUsers, userID)
	_, err := f.pool.Exec(context.Background(), `
		INSERT INTO usuario (id_usuario, username, correo, contrasena_hash)
		VALUES ($1, $2, $3, $4)`,
		userID,
		username,
		fmt.Sprintf("%s@example.test", userID),
		"test-password-hash",
	)
	require.NoError(t, err)
	return userID
}

func (f *fixture) createFriendship(t *testing.T, userA, userB int, status string) string {
	t.Helper()
	return f.createFriendshipIDs(t, f.userIDs[userA], f.userIDs[userB], status)
}

func (f *fixture) createFriendshipIDs(t *testing.T, userA, userB, status string) string {
	t.Helper()
	f.friendship++
	id := time.Now().UnixNano()%2_000_000_000 + f.friendship
	friendshipID := fmt.Sprintf("%d", id)
	var acceptedAt *time.Time
	if status == "ACEPTADA" {
		now := time.Now().UTC()
		acceptedAt = &now
	}
	_, err := f.pool.Exec(context.Background(), `
		INSERT INTO amistad (id_amistad, id_usuario_1, id_usuario_2, estado, fecha_amistad)
		VALUES ($1, $2, $3, $4, $5)`,
		friendshipID, userA, userB, status, acceptedAt)
	require.NoError(t, err)
	return friendshipID
}

func (f *fixture) setUserActivity(t *testing.T, user int, creation, lastLogin string) {
	t.Helper()
	_, err := f.pool.Exec(context.Background(), `
		UPDATE usuario
		SET fecha_creacion = CASE WHEN $2 = 'now' THEN CURRENT_TIMESTAMP
		                          WHEN $2 = 'null' THEN NULL
		                          ELSE NULL END,
		    fecha_ultimo_login = CASE WHEN $3 = 'now' THEN CURRENT_TIMESTAMP
		                              WHEN $3 = 'threshold' THEN DATE_TRUNC('day', CURRENT_TIMESTAMP) - INTERVAL '6 months'
		                              WHEN $3 = 'inactive' THEN CURRENT_TIMESTAMP - INTERVAL '7 months'
		                              ELSE NULL END
		WHERE id_usuario = $1`,
		f.userIDs[user], creation, lastLogin)
	require.NoError(t, err)
}

func (f *fixture) addPostActivity(t *testing.T, user int, age string) {
	t.Helper()
	runID := time.Now().UnixNano()
	postID := fmt.Sprintf("social-post-%d", runID)
	_, err := f.pool.Exec(context.Background(), `
		INSERT INTO publicacion (id_publicacion, id_usuario, texto, fecha_creacion)
		VALUES ($1, $2, 'activity test',
			CASE WHEN $3 = 'active' THEN CURRENT_TIMESTAMP
			     WHEN $3 = 'threshold' THEN CURRENT_TIMESTAMP - INTERVAL '6 months'
			     ELSE CURRENT_TIMESTAMP - INTERVAL '7 months' END)`,
		postID, f.userIDs[user], age)
	require.NoError(t, err)
}

func (f *fixture) addInteractionActivity(t *testing.T, user int, age string) {
	t.Helper()
	runID := time.Now().UnixNano()
	postID := fmt.Sprintf("social-interaction-post-%d", runID)
	interactionID := fmt.Sprintf("social-interaction-%d", runID)
	_, err := f.pool.Exec(context.Background(), `
		INSERT INTO publicacion (id_publicacion, id_usuario, texto, fecha_creacion)
		VALUES ($1, $2, 'interaction activity', CURRENT_TIMESTAMP)`,
		postID, f.userIDs[0])
	require.NoError(t, err)
	_, err = f.pool.Exec(context.Background(), `
		INSERT INTO interaccion (id_interaccion, id_usuario, id_publicacion, tipo, fecha_creacion)
		VALUES ($1, $2, $3, 'like',
			CASE WHEN $4 = 'active' THEN CURRENT_TIMESTAMP
			     WHEN $4 = 'threshold' THEN CURRENT_TIMESTAMP - INTERVAL '6 months'
			     ELSE CURRENT_TIMESTAMP - INTERVAL '7 months' END)`,
		interactionID, f.userIDs[user], postID, age)
	require.NoError(t, err)
}
