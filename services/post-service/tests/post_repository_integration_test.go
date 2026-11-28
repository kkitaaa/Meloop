package postservicetests

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/meloop/post-service/repositories"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openTestPostgres(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}

	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	require.NoError(t, err)
	db, err := gormDB.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, db.PingContext(ctx))
	return db
}

func TestPostgresPostLifecyclePersistsPostAndInteractions(t *testing.T) {
	db := openTestPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	authorID, readerID := uuid.NewString(), uuid.NewString()
	for _, id := range []string{authorID, readerID} {
		_, err := db.ExecContext(ctx, `
			INSERT INTO usuario (id_usuario, username, correo, contrasena_hash)
			VALUES ($1, $2, $3, $4)
		`, id, "post-test-"+id[:8], id+"@example.test", "integration-test")
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		_, err := db.Exec(`DELETE FROM usuario WHERE id_usuario IN ($1, $2)`, authorID, readerID)
		require.NoError(t, err)
	})

	repository := repositories.NewPostRepository(db)
	mediaURL := "https://cdn.example.test/post.png"
	post, err := repository.CreatePost(ctx, authorID, "persisted post", &mediaURL)
	require.NoError(t, err)
	require.NotEmpty(t, post.ID)
	_, err = uuid.Parse(post.ID)
	require.NoError(t, err)

	var persistedAuthorID, persistedContent, persistedMediaURL string
	err = db.QueryRowContext(ctx, `
		SELECT id_usuario, texto, url_multimedia
		FROM publicacion WHERE id_publicacion = $1
	`, post.ID).Scan(&persistedAuthorID, &persistedContent, &persistedMediaURL)
	require.NoError(t, err)
	require.Equal(t, authorID, persistedAuthorID)
	require.Equal(t, "persisted post", persistedContent)
	require.Equal(t, mediaURL, persistedMediaURL)

	require.NoError(t, repository.AddLike(ctx, post.ID, readerID))
	require.ErrorIs(t, repository.AddLike(ctx, post.ID, readerID), repositories.ErrAlreadyLiked)
	require.NoError(t, repository.SavePost(ctx, readerID, post.ID))
	require.NoError(t, repository.SavePost(ctx, readerID, post.ID))

	var likeCount, saveCount int
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT COUNT(*) FILTER (WHERE tipo = 'like'), COUNT(*) FILTER (WHERE tipo = 'save')
		FROM interaccion WHERE id_publicacion = $1 AND id_usuario = $2
	`, post.ID, readerID).Scan(&likeCount, &saveCount))
	require.Equal(t, 1, likeCount)
	require.Equal(t, 1, saveCount)

	savedPosts, err := repository.GetSavedPosts(ctx, readerID, nil, 10, 0)
	require.NoError(t, err)
	require.Len(t, savedPosts, 1)
	require.Equal(t, post.ID, savedPosts[0].ID)

	require.NoError(t, repository.UpdatePost(ctx, post.ID, authorID, "edited post"))
	err = db.QueryRowContext(ctx, `SELECT texto FROM publicacion WHERE id_publicacion = $1`, post.ID).Scan(&persistedContent)
	require.NoError(t, err)
	require.Equal(t, "edited post", persistedContent)
}
