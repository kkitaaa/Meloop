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
			SELECT id_usuario_2 AS friend_id
			FROM AMISTAD
			WHERE id_usuario_1 = ? AND UPPER(TRIM(estado)) = 'ACEPTADA'
			UNION
			SELECT id_usuario_1 AS friend_id
			FROM AMISTAD
			WHERE id_usuario_2 = ? AND UPPER(TRIM(estado)) = 'ACEPTADA'
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
			LEFT JOIN CONFIGURACION_PRIVACIDAD privacy
				ON privacy.id_usuario = p.id_usuario
			WHERE (
				p.id_usuario = ?
				OR (
					p.id_usuario IN (SELECT friend_id FROM friend_ids)
					AND UPPER(TRIM(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO'))) = 'AMIGOS'
				)
				OR UPPER(TRIM(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO'))) = 'PUBLICO'
			)
			AND NOT EXISTS (
				SELECT 1
				FROM BLOQUEO b
				WHERE (
					b.id_usuario_bloqueador = ?
					AND b.id_usuario_bloqueado = p.id_usuario
				) OR (
					b.id_usuario_bloqueado = ?
					AND b.id_usuario_bloqueador = p.id_usuario
				)
			)
		),
		likes_agg AS (
			SELECT
				id_publicacion,
				COUNT(DISTINCT id_usuario) AS likes_count,
				BOOL_OR(id_usuario = ?) AS is_liked
			FROM INTERACCION
			WHERE LOWER(TRIM(tipo)) = 'like'
			  AND id_publicacion IN (SELECT id_publicacion FROM candidate_posts)
			GROUP BY id_publicacion
		),
		comments_agg AS (
			SELECT
				id_publicacion,
				COUNT(id_interaccion) AS comments_count
			FROM INTERACCION
			WHERE LOWER(TRIM(tipo)) = 'comentario'
			  AND id_publicacion IN (SELECT id_publicacion FROM candidate_posts)
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
				COALESCE(l.is_liked, FALSE) AS is_liked,
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
										 AND can.id_cancion IS NOT NULL
										 AND EXISTS (
											SELECT 1
											FROM PREFERENCIA_MUSICAL author_pref
											WHERE author_pref.id_usuario = cp.id_usuario
											  AND UPPER(TRIM(author_pref.tipo)) = 'GENERO'
											  AND LOWER(TRIM(author_pref.spotify_id)) = LOWER(TRIM(pref.spotify_id))
										 )
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
			is_liked,
			comments_count,
			relevance_score
		FROM scored_posts
		ORDER BY
			CASE WHEN ? = 'relevant' THEN relevance_score END DESC,
			created_at DESC,
			id DESC
		LIMIT ? OFFSET ?
	`

	err := r.db.Raw(
		query,
		userID,
		userID,
		userID,
		userID,
		userID,
		userID,
		userID,
		algorithm,
		limit,
		offset,
	).Scan(&feed).Error
	if err != nil {
		return nil, err
	}

	return feed, nil
}
