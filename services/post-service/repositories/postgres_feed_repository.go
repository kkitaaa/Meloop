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
	        '' as author_avatar,
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
	    LEFT JOIN CONFIGURACION_PRIVACIDAD privacy
	        ON privacy.id_usuario = p.id_usuario
	    LEFT JOIN (
	        SELECT id_publicacion, COUNT(id_interaccion) as likes_count
	        FROM INTERACCION WHERE tipo = 'like' GROUP BY id_publicacion
	    ) l ON p.id_publicacion = l.id_publicacion
	    LEFT JOIN (
	        SELECT id_publicacion, COUNT(id_interaccion) as comments_count
	        FROM INTERACCION WHERE tipo = 'comentario' GROUP BY id_publicacion
	    ) c ON p.id_publicacion = c.id_publicacion
	    WHERE (
	        p.id_usuario = ?
	        OR (
	            p.id_usuario IN (
	                SELECT id_usuario_2 FROM AMISTAD
	                WHERE id_usuario_1 = ?
	                    AND UPPER(estado) = 'ACEPTADA'
	                UNION
	                SELECT id_usuario_1 FROM AMISTAD
	                WHERE id_usuario_2 = ?
	                    AND UPPER(estado) = 'ACEPTADA'
	            )
	            AND UPPER(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO')) = 'AMIGOS'
	        )
	        OR UPPER(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO')) = 'PUBLICO'
	    )
	    AND NOT EXISTS (
	        SELECT 1 FROM BLOQUEO b
	        WHERE (
	            b.id_usuario_bloqueador = ?
	            AND b.id_usuario_bloqueado = p.id_usuario
	        ) OR (
	            b.id_usuario_bloqueado = ?
	            AND b.id_usuario_bloqueador = p.id_usuario
	        )
	    )
	    ORDER BY p.fecha_creacion DESC
	    LIMIT ? OFFSET ?
	`

	// Mostramos publicaciones propias, de amistades aceptadas y públicas, excepto las de usuarios bloqueados.
	err := r.db.Raw(
		query,
		userID,
		userID,
		userID,
		userID,
		userID,
		userID,
		limit,
		offset,
	).Scan(&feed).Error
	if err != nil {
		return nil, err
	}

	return feed, nil
}
