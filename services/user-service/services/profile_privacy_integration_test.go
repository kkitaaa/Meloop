package services_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/repositories"
	"github.com/meloop/user-service/services"
	"github.com/stretchr/testify/require"
)

func TestThirdPartyProfilePrivacyPostgresIntegration(t *testing.T) {
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
	ownerID := "profile-privacy-owner-" + runID
	friendID := "profile-privacy-friend-" + runID
	strangerID := "profile-privacy-stranger-" + runID
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
			fmt.Sprintf("profileprivacy%s%d", runID, i),
			fmt.Sprintf("profileprivacy%s%d@example.test", runID, i),
			"test-password-hash",
		)
		require.NoError(t, err)
	}
	_, err = pool.Exec(context.Background(), `
		INSERT INTO amistad (id_amistad, id_usuario_1, id_usuario_2, estado)
		VALUES ($1, $2, $3, 'ACEPTADA')`,
		"profile-privacy-friendship-"+runID, ownerID, friendID,
	)
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(), `
		INSERT INTO publicacion (id_publicacion, id_usuario, texto)
		VALUES ($1, $2, 'profile privacy integration post')`,
		"profile-privacy-post-"+runID, ownerID,
	)
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(), `
		INSERT INTO configuracion_privacidad (
			id_privacidad, id_usuario, visibilidad_perfil, visibilidad_publicaciones,
			visibilidad_interacciones, permitir_mensajes, permitir_solicitudes_amistad
		) VALUES (gen_random_uuid()::text, $1, 'PUBLICO', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'TODOS')`,
		ownerID,
	)
	require.NoError(t, err)

	profileService := services.NewProfileService(
		repositories.NewProfileRepository(pool),
		repositories.NewUserRepository(pool),
		nil,
		repositories.NewPrivacyRepository(pool),
	)

	friendProfile, err := profileService.GetThirdPartyProfile(context.Background(), friendID, ownerID)
	require.NoError(t, err)
	require.False(t, friendProfile.PerfilPrivado)
	require.EqualValues(t, 1, friendProfile.Estadisticas.Publicaciones)
	require.EqualValues(t, 1, friendProfile.Estadisticas.Amigos)

	strangerProfile, err := profileService.GetThirdPartyProfile(context.Background(), strangerID, ownerID)
	require.NoError(t, err)
	require.False(t, strangerProfile.PerfilPrivado)
	require.EqualValues(t, 0, strangerProfile.Estadisticas.Publicaciones)
	require.EqualValues(t, 0, strangerProfile.Estadisticas.Amigos)

	_, err = pool.Exec(context.Background(), `
		UPDATE configuracion_privacidad SET visibilidad_perfil = 'AMIGOS'
		WHERE id_usuario = $1`, ownerID)
	require.NoError(t, err)

	strangerProfile, err = profileService.GetThirdPartyProfile(context.Background(), strangerID, ownerID)
	require.NoError(t, err)
	require.True(t, strangerProfile.PerfilPrivado)
	require.Nil(t, strangerProfile.Estadisticas)
}
