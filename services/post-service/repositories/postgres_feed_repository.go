package repositories

import (
	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
)

type FeedRepository interface {
	GetMainFeed(userID string, limit, offset int, algorithm string) ([]models.FeedPost, error)
}

type PostgresFeedRepository struct {
	db *gorm.DB
}

func NewPostgresFeedRepository(db *gorm.DB) *PostgresFeedRepository {
	return &PostgresFeedRepository{db: db}
}

func (r *PostgresFeedRepository) GetMainFeed(userID string, limit, offset int, algorithm string) ([]models.FeedPost, error) {
	var feed []models.FeedPost

	// Consulta SQL optimizada:
	// 1. Pre-agrega likes y comentarios para no asfixiar la BD (RF-30).
	// 2. Cruza con CANCION y PREFERENCIA_MUSICAL para detectar afinidad.
	// 3. Ordena dinámicamente por relevancia (popularidad + gustos) o por fecha.
	query := `
		SELECT 
			p.id_publicacion as id, 
			p.texto as content, 
			p.url_multimedia as media_url, 
			p.fecha_creacion as created_at,
			u.id_usuario as author_id, 
			u.username as author_username, 
			'' as author_avatar, 
			COALESCE(l.likes_count, 0) as likes_count,
			COALESCE(c.comments_count, 0) as comments_count,
			(
				(COALESCE(l.likes_count, 0) * 2) + 
				(COALESCE(c.comments_count, 0) * 3) +
				(CASE WHEN pref.id_usuario IS NOT NULL THEN 50 ELSE 0 END)
			) as relevance_score
		FROM PUBLICACION p
		JOIN USUARIO u ON p.id_usuario = u.id_usuario
		LEFT JOIN CANCION can ON p.id_cancion = can.id_cancion
		LEFT JOIN (
			SELECT id_publicacion, COUNT(id_interaccion) as likes_count 
			FROM INTERACCION WHERE tipo = 'like' GROUP BY id_publicacion
		) l ON p.id_publicacion = l.id_publicacion
		LEFT JOIN (
			SELECT id_publicacion, COUNT(id_interaccion) as comments_count 
			FROM INTERACCION WHERE tipo = 'comentario' GROUP BY id_publicacion
		) c ON p.id_publicacion = c.id_publicacion
		LEFT JOIN PREFERENCIA_MUSICAL pref 
			ON pref.id_usuario = ? AND pref.spotify_id = can.spotify_id
		WHERE (
			p.id_usuario IN (
				SELECT id_usuario_2 FROM AMISTAD WHERE id_usuario_1 = ? AND estado = 'aceptada'
				UNION
				SELECT id_usuario_1 FROM AMISTAD WHERE id_usuario_2 = ? AND estado = 'aceptada'
			) 
		)
		ORDER BY 
			CASE WHEN ? = 'relevant' THEN (
				(COALESCE(l.likes_count, 0) * 2) + 
				(COALESCE(c.comments_count, 0) * 3) +
				(CASE WHEN pref.id_usuario IS NOT NULL THEN 50 ELSE 0 END)
			) END DESC,
			p.fecha_creacion DESC
		LIMIT ? OFFSET ?
	`

	// Parámetros: userID (preferencias), userID x2 (amistad), algorithm (para el order by), limit, offset
	err := r.db.Raw(query, userID, userID, userID, algorithm, limit, offset).Scan(&feed).Error
	if err != nil {
		return nil, err
	}

	return feed, nil
}
