package repositories

import (
	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
)

type FeedRepository interface {
	GetMainFeed(userID string, limit, offset int) ([]models.FeedPost, error)
}

type PostgresFeedRepository struct {
	db *gorm.DB
}

func NewPostgresFeedRepository(db *gorm.DB) *PostgresFeedRepository {
	return &PostgresFeedRepository{db: db}
}

func (r *PostgresFeedRepository) GetMainFeed(userID string, limit, offset int) ([]models.FeedPost, error) {
	var feed []models.FeedPost

	// Consulta SQL basada en el diagrama ER exacto.
	// Nota: Agrupamos las interacciones por 'tipo' para contar likes y comentarios.
	query := `
		SELECT 
			p.id_publicacion as id, 
			p.texto as content, 
			p.url_multimedia as media_url, 
			p.fecha_creacion as created_at,
			u.id_usuario as author_id, 
			u.username as author_username, 
			'' as author_avatar, -- Placeholder: No hay avatar en la tabla USUARIO actual
			COALESCE(l.likes_count, 0) as likes_count,
			EXISTS (
				SELECT 1 FROM INTERACCION il
				WHERE il.id_publicacion = p.id_publicacion
					AND il.id_usuario = ?
					AND il.tipo = 'like'
			) as is_liked,
			COALESCE(c.comments_count, 0) as comments_count
		FROM PUBLICACION p
		JOIN USUARIO u ON p.id_usuario = u.id_usuario
		LEFT JOIN (
			SELECT id_publicacion, COUNT(id_interaccion) as likes_count 
			FROM INTERACCION WHERE tipo = 'like' GROUP BY id_publicacion
		) l ON p.id_publicacion = l.id_publicacion
		LEFT JOIN (
			SELECT id_publicacion, COUNT(id_interaccion) as comments_count 
			FROM INTERACCION WHERE tipo = 'comentario' GROUP BY id_publicacion
		) c ON p.id_publicacion = c.id_publicacion
		WHERE meloop_privacy_allows_visibility(
			p.id_usuario::text,
			?::text,
			(SELECT cp.visibilidad_publicaciones
			 FROM CONFIGURACION_PRIVACIDAD cp
			 WHERE cp.id_usuario = p.id_usuario)
		)
		AND NOT EXISTS (
			SELECT 1 FROM BLOQUEO b
			WHERE (b.id_usuario_bloqueador = ? AND b.id_usuario_bloqueado = p.id_usuario)
			   OR (b.id_usuario_bloqueado = ? AND b.id_usuario_bloqueador = p.id_usuario)
		)
		ORDER BY p.fecha_creacion DESC
		LIMIT ? OFFSET ?
	`

	// Mostramos publicaciones permitidas por privacidad, excepto las de usuarios bloqueados.
	err := r.db.Raw(query, userID, userID, userID, userID, limit, offset).Scan(&feed).Error
	if err != nil {
		return nil, err
	}

	return feed, nil
}
