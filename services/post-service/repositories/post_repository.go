package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrPostNotFoundOrUnauthorized = errors.New("publicación no encontrada o no tienes permisos de autor")
	ErrModerationTargetNotFound   = errors.New("moderation target not found or not visible")
)

type ModerationTarget struct {
	AuthorID   string
	AuthorName string
	Text       string
	MediaURL   string
}

// Post representa una publicación basada estrictamente en el MER
type Post struct {
	ID            string    `json:"id"`
	AuthorID      string    `json:"author_id"`
	Content       string    `json:"content"`
	MusicID       *string   `json:"music_id,omitempty"`
	UrlMultimedia *string   `json:"url_multimedia,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// Cambiamos el nombre de la estructura a SQLPostRepository para evitar el conflicto
type SQLPostRepository struct {
	db *sql.DB
}

// Mantenemos NewPostRepository para no romper las inyecciones en main.go
func NewPostRepository(db *sql.DB) *SQLPostRepository {
	return &SQLPostRepository{db: db}
}

// UpdatePost modifica el contenido de una publicación usando el esquema real del MER
func (r *SQLPostRepository) UpdatePost(ctx context.Context, postID string, authorID string, newContent string) error {
	query := `
		UPDATE publicacion 
		SET texto = $1 
		WHERE id_publicacion = $2 AND id_usuario = $3
	`

	result, err := r.db.ExecContext(ctx, query, newContent, postID, authorID)
	if err != nil {
		return fmt.Errorf("error al actualizar la publicación: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return ErrPostNotFoundOrUnauthorized
	}

	return nil
}

// DeletePost elimina una publicación validando la autoría con el MER
func (r *SQLPostRepository) DeletePost(ctx context.Context, postID string, authorID string) error {
	query := `
		DELETE FROM publicacion 
		WHERE id_publicacion = $1 AND id_usuario = $2
	`

	result, err := r.db.ExecContext(ctx, query, postID, authorID)
	if err != nil {
		return fmt.Errorf("error al eliminar la publicación: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return ErrPostNotFoundOrUnauthorized
	}

	return nil
}

// GetPostByID obtiene un post por su ID
func (r *SQLPostRepository) GetPostByID(ctx context.Context, postID string, requesterID string) (*Post, error) {
	query := `
		SELECT id_publicacion, id_usuario, texto, id_cancion, url_multimedia, fecha_creacion
		FROM publicacion
		WHERE id_publicacion = $1
	`
	row := r.db.QueryRowContext(ctx, query, postID)

	var post Post
	err := row.Scan(&post.ID, &post.AuthorID, &post.Content, &post.MusicID, &post.UrlMultimedia, &post.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("publicación no encontrada")
		}
		return nil, fmt.Errorf("error al obtener post: %w", err)
	}

	return &post, nil
}

func (r *SQLPostRepository) GetModerationTarget(
	ctx context.Context,
	targetType string,
	targetID string,
	requesterID string,
) (*ModerationTarget, error) {
	var query string
	switch targetType {
	case "post":
		query = `
			SELECT p.id_usuario::text, u.username, COALESCE(p.texto, ''), COALESCE(p.url_multimedia, '')
			FROM publicacion p
			JOIN usuario u ON u.id_usuario = p.id_usuario
			LEFT JOIN configuracion_privacidad privacy ON privacy.id_usuario = p.id_usuario
			WHERE p.id_publicacion::text = $1
			  AND (
				p.id_usuario::text = $2
				OR UPPER(COALESCE(NULLIF(BTRIM(privacy.visibilidad_publicaciones), ''), 'PUBLICO')) = 'PUBLICO'
				OR (
					UPPER(COALESCE(NULLIF(BTRIM(privacy.visibilidad_publicaciones), ''), 'PUBLICO')) = 'AMIGOS'
					AND EXISTS (
						SELECT 1 FROM amistad a
						WHERE UPPER(a.estado) = 'ACEPTADA'
						  AND ((a.id_usuario_1 = p.id_usuario AND a.id_usuario_2::text = $2)
						    OR (a.id_usuario_2 = p.id_usuario AND a.id_usuario_1::text = $2))
					)
				)
			  )
			  AND NOT EXISTS (
				SELECT 1 FROM bloqueo b
				WHERE (b.id_usuario_bloqueador::text = $2 AND b.id_usuario_bloqueado = p.id_usuario)
				   OR (b.id_usuario_bloqueado::text = $2 AND b.id_usuario_bloqueador = p.id_usuario)
			  )
		`
	case "comment":
		query = `
			SELECT c.id_usuario::text, comment_author.username, COALESCE(c.texto, ''), COALESCE(p.url_multimedia, '')
			FROM comentario c
			JOIN publicacion p ON p.id_publicacion = c.id_publicacion
			JOIN usuario comment_author ON comment_author.id_usuario = c.id_usuario
			LEFT JOIN configuracion_privacidad post_privacy ON post_privacy.id_usuario = p.id_usuario
			LEFT JOIN configuracion_privacidad comment_privacy ON comment_privacy.id_usuario = c.id_usuario
			WHERE c.id_comentario::text = $1
			  AND (
				p.id_usuario::text = $2
				OR UPPER(COALESCE(NULLIF(BTRIM(post_privacy.visibilidad_publicaciones), ''), 'PUBLICO')) = 'PUBLICO'
				OR (
					UPPER(COALESCE(NULLIF(BTRIM(post_privacy.visibilidad_publicaciones), ''), 'PUBLICO')) = 'AMIGOS'
					AND EXISTS (
						SELECT 1 FROM amistad a
						WHERE UPPER(a.estado) = 'ACEPTADA'
						  AND ((a.id_usuario_1 = p.id_usuario AND a.id_usuario_2::text = $2)
						    OR (a.id_usuario_2 = p.id_usuario AND a.id_usuario_1::text = $2))
					)
				)
			  )
			  AND (
				c.id_usuario::text = $2
				OR UPPER(COALESCE(NULLIF(BTRIM(comment_privacy.visibilidad_interacciones), ''), 'PUBLICO')) = 'PUBLICO'
				OR (
					UPPER(COALESCE(NULLIF(BTRIM(comment_privacy.visibilidad_interacciones), ''), 'PUBLICO')) = 'AMIGOS'
					AND EXISTS (
						SELECT 1 FROM amistad a
						WHERE UPPER(a.estado) = 'ACEPTADA'
						  AND ((a.id_usuario_1 = c.id_usuario AND a.id_usuario_2::text = $2)
						    OR (a.id_usuario_2 = c.id_usuario AND a.id_usuario_1::text = $2))
					)
				)
			  )
			  AND NOT EXISTS (
				SELECT 1 FROM bloqueo b
				WHERE (b.id_usuario_bloqueador::text = $2 AND b.id_usuario_bloqueado IN (p.id_usuario, c.id_usuario))
				   OR (b.id_usuario_bloqueado::text = $2 AND b.id_usuario_bloqueador IN (p.id_usuario, c.id_usuario))
			  )
		`
	default:
		return nil, ErrModerationTargetNotFound
	}

	var target ModerationTarget
	if err := r.db.QueryRowContext(ctx, query, targetID, requesterID).Scan(
		&target.AuthorID, &target.AuthorName, &target.Text, &target.MediaURL,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrModerationTargetNotFound
		}
		return nil, fmt.Errorf("get visible moderation target: %w", err)
	}
	return &target, nil
}

// GetPostsByAuthor obtiene los posts de un perfil con paginación
func (r *SQLPostRepository) GetPostsByAuthor(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]Post, error) {
	query := `
		SELECT id_publicacion, id_usuario, texto, id_cancion, url_multimedia, fecha_creacion
		FROM publicacion
		WHERE id_usuario = $1 
		ORDER BY fecha_creacion DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al consultar posts por autor: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Content, &p.MusicID, &p.UrlMultimedia, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	if posts == nil {
		posts = []Post{}
	}
	return posts, nil
}

// SearchPosts busca posts por texto o canción
func (r *SQLPostRepository) SearchPosts(ctx context.Context, searchTerm string, requesterID string, limit int, offset int) ([]Post, error) {
	searchPattern := "%" + searchTerm + "%"

	query := `
		SELECT id_publicacion, id_usuario, texto, id_cancion, url_multimedia, fecha_creacion
		FROM publicacion
		WHERE (texto ILIKE $1 OR id_cancion ILIKE $1)
		ORDER BY fecha_creacion DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al buscar posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Content, &p.MusicID, &p.UrlMultimedia, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	if posts == nil {
		posts = []Post{}
	}
	return posts, nil
}

// CreatePost inserta una nueva publicación en la base de datos
func (r *SQLPostRepository) CreatePost(ctx context.Context, authorID string, content string, urlMultimedia *string) (*Post, error) {
	// Generamos un ID simple basado en el timestamp actual (puedes cambiarlo a UUID si tu BD lo usa)
	newID := fmt.Sprintf("post-%d", time.Now().UnixMilli())

	query := `
		INSERT INTO publicacion (id_publicacion, id_usuario, texto, url_multimedia, fecha_creacion)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id_publicacion, id_usuario, texto, url_multimedia, fecha_creacion
	`
	var post Post
	err := r.db.QueryRowContext(ctx, query, newID, authorID, content, urlMultimedia).Scan(
		&post.ID, &post.AuthorID, &post.Content, &post.UrlMultimedia, &post.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error insertando publicación: %w", err)
	}

	return &post, nil
}
