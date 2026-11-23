package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/meloop/post-service/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostRepositoriesEnforcePrivacyPostgres(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL no está configurada; se omite la prueba de integración")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(config.ConnConfig.Database), "test",
		"TEST_DATABASE_URL debe apuntar a una base de datos cuyo nombre incluya 'test'")

	ctx := context.Background()
	sqlDB, err := sql.Open("pgx", databaseURL)
	require.NoError(t, err)
	require.NoError(t, sqlDB.PingContext(ctx))

	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	require.NoError(t, err)
	gormSQLDB, err := gormDB.DB()
	require.NoError(t, err)

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	authorID := "post-privacy-author-" + runID
	friendID := "post-privacy-friend-" + runID
	strangerID := "post-privacy-stranger-" + runID
	postID := "privacy-post-" + runID
	searchText := "privacy-search-" + runID
	userIDs := []string{authorID, friendID, strangerID}
	t.Cleanup(func() {
		_, cleanupErr := sqlDB.ExecContext(ctx, `DELETE FROM usuario WHERE id_usuario = ANY($1)`, userIDs)
		_ = gormSQLDB.Close()
		_ = sqlDB.Close()
		require.NoError(t, cleanupErr)
	})

	for i, userID := range userIDs {
		_, err := sqlDB.ExecContext(ctx, `
			INSERT INTO usuario (id_usuario, username, correo, contrasena_hash)
			VALUES ($1, $2, $3, $4)`,
			userID,
			fmt.Sprintf("postprivacy%s%d", runID, i),
			fmt.Sprintf("postprivacy%s%d@example.test", runID, i),
			"test-password-hash",
		)
		require.NoError(t, err)
	}
	_, err = sqlDB.ExecContext(ctx, `
		INSERT INTO configuracion_privacidad (
			id_privacidad, id_usuario, visibilidad_perfil, visibilidad_publicaciones,
			visibilidad_interacciones, permitir_mensajes, permitir_solicitudes_amistad
		) VALUES (gen_random_uuid()::text, $1, 'PUBLICO', 'AMIGOS', 'PUBLICO', 'AMIGOS', 'TODOS')`,
		authorID,
	)
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, `
		INSERT INTO amistad (id_amistad, id_usuario_1, id_usuario_2, estado)
		VALUES ($1, $2, $3, 'ACEPTADA')`,
		"post-privacy-friendship-"+runID, authorID, friendID,
	)
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, `
		INSERT INTO publicacion (id_publicacion, id_usuario, texto, fecha_creacion)
		VALUES ($1, $2, $3, NOW())`,
		postID, authorID, searchText,
	)
	require.NoError(t, err)

	postRepo := NewPostRepository(sqlDB)
	feedRepo := NewPostgresFeedRepository(gormDB)

	_, err = postRepo.GetPostByID(ctx, postID, friendID)
	require.NoError(t, err)
	_, err = postRepo.GetPostByID(ctx, postID, strangerID)
	require.Error(t, err)

	friendPosts, err := postRepo.GetPostsByAuthor(ctx, authorID, friendID, 10, 0)
	require.NoError(t, err)
	require.Len(t, friendPosts, 1)
	strangerPosts, err := postRepo.GetPostsByAuthor(ctx, authorID, strangerID, 10, 0)
	require.NoError(t, err)
	require.Empty(t, strangerPosts)

	friendSearch, err := postRepo.SearchPosts(ctx, searchText, friendID, 10, 0)
	require.NoError(t, err)
	require.Len(t, friendSearch, 1)
	strangerSearch, err := postRepo.SearchPosts(ctx, searchText, strangerID, 10, 0)
	require.NoError(t, err)
	require.Empty(t, strangerSearch)

	friendFeed, err := feedRepo.GetMainFeed(friendID, 10, 0)
	require.NoError(t, err)
	require.Contains(t, feedPostIDs(friendFeed), postID)
	strangerFeed, err := feedRepo.GetMainFeed(strangerID, 10, 0)
	require.NoError(t, err)
	require.NotContains(t, feedPostIDs(strangerFeed), postID)
}

func feedPostIDs(posts []models.FeedPost) []string {
	ids := make([]string, 0, len(posts))
	for _, post := range posts {
		ids = append(ids, post.ID)
	}
	return ids
}
