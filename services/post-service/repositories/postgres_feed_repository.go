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

	query := `
		WITH friend_ids AS (
			SELECT id_usuario_2 AS friend_id FROM AMISTAD WHERE id_usuario_1 = ? AND UPPER(TRIM(estado)) = 'ACEPTADA'
			UNION
			SELECT id_usuario_1 AS friend_id FROM AMISTAD WHERE id_usuario_2 = ? AND UPPER(TRIM(estado)) = 'ACEPTADA'
		),
		candidate_posts AS (
			SELECT 
				p.id_publicacion,
				p.texto,
				p.url_multimedia,
				p.fecha_creacion,
				p.id_usuario,
				p.id_cancion
			FROM PUBLICACION p
			WHERE p.id_usuario IN (SELECT friend_id FROM friend_ids)
		),
		likes_agg AS (
			SELECT id_publicacion, COUNT(*) AS likes_count
			FROM (
				SELECT id_usuario, id_publicacion
				FROM megusta
				WHERE id_publicacion IN (SELECT id_publicacion FROM candidate_posts)
				UNION
				SELECT id_usuario, id_publicacion
				FROM INTERACCION
				WHERE LOWER(TRIM(tipo)) = 'like'
				  AND id_publicacion IN (SELECT id_publicacion FROM candidate_posts)
			) all_likes
			GROUP BY id_publicacion
		),
		comments_agg AS (
			SELECT id_publicacion, COUNT(*) AS comments_count
			FROM (
				SELECT id_publicacion
				FROM comentario
				WHERE id_publicacion IN (SELECT id_publicacion FROM candidate_posts)
				UNION ALL
				SELECT id_publicacion
				FROM INTERACCION
				WHERE LOWER(TRIM(tipo)) = 'comentario'
				  AND id_publicacion IN (SELECT id_publicacion FROM candidate_posts)
			) all_comments
			GROUP BY id_publicacion
		),
		scored_posts AS (
			SELECT 
				cp.id_publicacion AS id,
				cp.texto AS content,
				cp.url_multimedia AS media_url,
				cp.fecha_creacion AS created_at,
				u.id_usuario AS author_id,
				u.username AS author_username,
				'' AS author_avatar,
				COALESCE(l.likes_count, 0) AS likes_count,
				COALESCE(c.comments_count, 0) AS comments_count,
				(
					(COALESCE(l.likes_count, 0) * 2) +
					(COALESCE(c.comments_count, 0) * 3) +
					COALESCE((
						SELECT SUM(match_score)
						FROM (
							SELECT DISTINCT
								UPPER(TRIM(pref.tipo)) AS pref_tipo,
								LOWER(TRIM(pref.spotify_id)) AS pref_valor,
								CASE
									WHEN UPPER(TRIM(pref.tipo)) = 'CANCION'
										 AND can.spotify_id IS NOT NULL
										 AND LOWER(TRIM(pref.spotify_id)) = LOWER(TRIM(can.spotify_id))
									THEN 50
									WHEN UPPER(TRIM(pref.tipo)) = 'ARTISTA'
										 AND can.artista_nombre IS NOT NULL
										 AND LOWER(TRIM(pref.spotify_id)) = LOWER(TRIM(can.artista_nombre))
									THEN 30
									WHEN UPPER(TRIM(pref.tipo)) = 'GENERO'
										 AND LENGTH(TRIM(pref.spotify_id)) >= 3
										 AND (' ' || regexp_replace(LOWER(COALESCE(cp.texto, '')), '[^[:alnum:]]+', ' ', 'g') || ' ')
										     LIKE ('% ' || regexp_replace(LOWER(TRIM(pref.spotify_id)), '[^[:alnum:]]+', ' ', 'g') || ' %')
									THEN 15
									ELSE 0
								END AS match_score
							FROM PREFERENCIA_MUSICAL pref
							WHERE pref.id_usuario = ?
						) matched_prefs
					), 0)
				) AS relevance_score
			FROM candidate_posts cp
			JOIN USUARIO u ON cp.id_usuario = u.id_usuario
			LEFT JOIN CANCION can ON cp.id_cancion = can.id_cancion
			LEFT JOIN likes_agg l ON cp.id_publicacion = l.id_publicacion
			LEFT JOIN comments_agg c ON cp.id_publicacion = c.id_publicacion
		)
		SELECT 
			id,
			content,
			media_url,
			created_at,
			author_id,
			author_username,
			author_avatar,
			likes_count,
			comments_count,
			relevance_score
		FROM scored_posts
		ORDER BY 
			CASE WHEN ? = 'relevant' THEN relevance_score END DESC,
			created_at DESC
		LIMIT ? OFFSET ?
	`

	err := r.db.Raw(query, userID, userID, userID, algorithm, limit, offset).Scan(&feed).Error
	if err != nil {
		return nil, err
	}

	return feed, nil
}