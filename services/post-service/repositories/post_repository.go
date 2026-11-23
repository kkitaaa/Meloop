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
	ErrAlreadyLiked               = errors.New("el usuario ya dio me gusta a esta publicación")
)

// Post representa una publicación basada estrictamente en el MER
type Post struct {
	ID            string    `json:"id"`
	AuthorID      string    `json:"author_id"`
	Content       string    `json:"content"`
	MusicID       *string   `json:"music_id,omitempty"`
	UrlMultimedia *string   `json:"url_multimedia,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// SQLPostRepository maneja la persistencia de las publicaciones
type SQLPostRepository struct {
	db *sql.DB
}

// NewPostRepository crea una nueva instancia del repositorio
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
		SELECT p.id_publicacion, p.id_usuario, p.texto, p.id_cancion, p.url_multimedia, p.fecha_creacion
		FROM publicacion p
		WHERE p.id_publicacion = $1
		  AND (
			p.id_usuario = $2
			OR COALESCE((
				SELECT UPPER(TRIM(cp.visibilidad_publicaciones))
				FROM configuracion_privacidad cp
				WHERE cp.id_usuario = p.id_usuario
			), 'PUBLICO') = 'PUBLICO'
			OR (
				COALESCE((
					SELECT UPPER(TRIM(cp.visibilidad_publicaciones))
					FROM configuracion_privacidad cp
					WHERE cp.id_usuario = p.id_usuario
				), 'PUBLICO') = 'AMIGOS'
				AND EXISTS (
					SELECT 1 FROM amistad a
					WHERE ((a.id_usuario_1 = p.id_usuario AND a.id_usuario_2 = $2)
					    OR (a.id_usuario_1 = $2 AND a.id_usuario_2 = p.id_usuario))
					  AND UPPER(a.estado) = 'ACEPTADA'
				)
			)
		  )
	`
	row := r.db.QueryRowContext(ctx, query, postID, requesterID)

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

// GetPostsByAuthor obtiene los posts de un perfil con paginación
func (r *SQLPostRepository) GetPostsByAuthor(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]Post, error) {
	query := `
		SELECT p.id_publicacion, p.id_usuario, p.texto, p.id_cancion, p.url_multimedia, p.fecha_creacion
		FROM publicacion p
		WHERE p.id_usuario = $1
		  AND (
			p.id_usuario = $2
			OR COALESCE((
				SELECT UPPER(TRIM(cp.visibilidad_publicaciones))
				FROM configuracion_privacidad cp
				WHERE cp.id_usuario = p.id_usuario
			), 'PUBLICO') = 'PUBLICO'
			OR (
				COALESCE((
					SELECT UPPER(TRIM(cp.visibilidad_publicaciones))
					FROM configuracion_privacidad cp
					WHERE cp.id_usuario = p.id_usuario
				), 'PUBLICO') = 'AMIGOS'
				AND EXISTS (
					SELECT 1 FROM amistad a
					WHERE ((a.id_usuario_1 = p.id_usuario AND a.id_usuario_2 = $2)
					    OR (a.id_usuario_1 = $2 AND a.id_usuario_2 = p.id_usuario))
					  AND UPPER(a.estado) = 'ACEPTADA'
				)
			)
		  )
		ORDER BY fecha_creacion DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, authorID, requesterID, limit, offset)
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
		SELECT p.id_publicacion, p.id_usuario, p.texto, p.id_cancion, p.url_multimedia, p.fecha_creacion
		FROM publicacion p
		WHERE (p.texto ILIKE $1 OR p.id_cancion ILIKE $1)
		  AND (
			p.id_usuario = $2
			OR COALESCE((
				SELECT UPPER(TRIM(cp.visibilidad_publicaciones))
				FROM configuracion_privacidad cp
				WHERE cp.id_usuario = p.id_usuario
			), 'PUBLICO') = 'PUBLICO'
			OR (
				COALESCE((
					SELECT UPPER(TRIM(cp.visibilidad_publicaciones))
					FROM configuracion_privacidad cp
					WHERE cp.id_usuario = p.id_usuario
				), 'PUBLICO') = 'AMIGOS'
				AND EXISTS (
					SELECT 1 FROM amistad a
					WHERE ((a.id_usuario_1 = p.id_usuario AND a.id_usuario_2 = $2)
					    OR (a.id_usuario_1 = $2 AND a.id_usuario_2 = p.id_usuario))
					  AND UPPER(a.estado) = 'ACEPTADA'
				)
			)
		  )
		ORDER BY fecha_creacion DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, searchPattern, requesterID, limit, offset)
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

// AddLike inserta una interacción de tipo 'like' validando que no exista previamente (RF-20)
func (r *SQLPostRepository) AddLike(ctx context.Context, postID string, userID string) error {
	// 1. Verificamos si ya existe para evitar duplicados
	checkQuery := `SELECT COUNT(1) FROM interaccion WHERE id_usuario = $1 AND id_publicacion = $2 AND tipo = 'like'`
	var count int
	if err := r.db.QueryRowContext(ctx, checkQuery, userID, postID).Scan(&count); err != nil {
		return fmt.Errorf("error al verificar el estado del like: %w", err)
	}

	if count > 0 {
		return ErrAlreadyLiked
	}

	// 2. Generamos un ID simple para la interacción y la insertamos
	interaccionID := fmt.Sprintf("like-%s-%d", userID, time.Now().UnixMilli())

	query := `
		INSERT INTO interaccion (id_interaccion, id_usuario, id_publicacion, tipo, fecha_creacion)
		VALUES ($1, $2, $3, 'like', NOW())
	`

	_, err := r.db.ExecContext(ctx, query, interaccionID, userID, postID)
	if err != nil {
		return fmt.Errorf("error al registrar el like: %w", err)
	}

	return nil
}

// RemoveLike elimina una interacción de tipo 'like' de la base de datos (RF-20)
func (r *SQLPostRepository) RemoveLike(ctx context.Context, postID string, userID string) error {
	query := `
		DELETE FROM interaccion 
		WHERE id_publicacion = $1 AND id_usuario = $2 AND tipo = 'like'
	`

	_, err := r.db.ExecContext(ctx, query, postID, userID)
	if err != nil {
		return fmt.Errorf("error al eliminar el like: %w", err)
	}

	return nil
}
