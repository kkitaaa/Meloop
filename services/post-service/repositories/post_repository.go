package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrPostNotFound = errors.New("publicación no encontrada")
	ErrUnauthorized = errors.New("no tienes permisos para modificar esta publicación")
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

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// verifyOwnership es un helper interno para validar seguridad antes de modificar
func (r *PostRepository) verifyOwnership(ctx context.Context, postID string, authorID string) error {
	var ownerID string
	query := `SELECT id_usuario FROM publicacion WHERE id_publicacion = $1`

	err := r.db.QueryRowContext(ctx, query, postID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPostNotFound
		}
		return fmt.Errorf("error al verificar publicación: %w", err)
	}

	if ownerID != authorID {
		return ErrUnauthorized
	}
	return nil
}

// UpdatePost modifica el contenido de una publicación usando el esquema real del MER
func (r *PostRepository) UpdatePost(ctx context.Context, postID string, authorID string, newContent string) error {
	if err := r.verifyOwnership(ctx, postID, authorID); err != nil {
		return err
	}

	query := `
		UPDATE publicacion 
		SET texto = $1 
		WHERE id_publicacion = $2
	`
	_, err := r.db.ExecContext(ctx, query, newContent, postID)
	if err != nil {
		return fmt.Errorf("error al actualizar la publicación: %w", err)
	}

	return nil
}

// DeletePost elimina una publicación validando la autoría con el MER
func (r *PostRepository) DeletePost(ctx context.Context, postID string, authorID string) error {
	if err := r.verifyOwnership(ctx, postID, authorID); err != nil {
		return err
	}

	query := `DELETE FROM publicacion WHERE id_publicacion = $1`

	_, err := r.db.ExecContext(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("error al eliminar la publicación: %w", err)
	}

	return nil
}

// GetPostByID obtiene un post por su ID
func (r *PostRepository) GetPostByID(ctx context.Context, postID string, requesterID string) (*Post, error) {
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
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("error al obtener post: %w", err)
	}

	return &post, nil
}

// GetPostsByAuthor obtiene los posts de un perfil con paginación
func (r *PostRepository) GetPostsByAuthor(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]Post, error) {
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
func (r *PostRepository) SearchPosts(ctx context.Context, searchTerm string, requesterID string, limit int, offset int) ([]Post, error) {
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
