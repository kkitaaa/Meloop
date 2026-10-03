package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Post representa una publicación en la base de datos
type Post struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	MusicID   *string   `json:"music_id,omitempty"`
	IsPrivate bool      `json:"is_private"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// GetPostByID obtiene un post validando que el autor no esté bloqueado por el solicitante
func (r *PostRepository) GetPostByID(ctx context.Context, postID string, requesterID string) (*Post, error) {
	// Simulamos la validación de bloqueo con una subquery (asumiendo tabla user_blocks)
	query := `
		SELECT p.id, p.author_id, p.content, p.music_id, p.is_private, p.created_at, p.updated_at
		FROM posts p
		WHERE p.id = $1 
		AND p.author_id NOT IN (
			SELECT blocked_id FROM user_blocks WHERE blocker_id = $2
		)
	`
	
	row := r.db.QueryRowContext(ctx, query, postID, requesterID)
	
	var post Post
	err := row.Scan(&post.ID, &post.AuthorID, &post.Content, &post.MusicID, &post.IsPrivate, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("publicación no encontrada o acceso denegado")
		}
		return nil, fmt.Errorf("error al obtener post: %w", err)
	}

	return &post, nil
}

// GetPostsByAuthor obtiene los posts de un perfil con paginación (RF-19)
func (r *PostRepository) GetPostsByAuthor(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]Post, error) {
	query := `
		SELECT id, author_id, content, music_id, is_private, created_at, updated_at
		FROM posts
		WHERE author_id = $1 
		AND (is_private = false OR author_id = $2)
		ORDER BY created_at DESC
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
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Content, &p.MusicID, &p.IsPrivate, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	
	if posts == nil {
		posts = []Post{} // Para devolver un JSON [] en lugar de null
	}
	return posts, nil
}

// SearchPosts busca posts por texto (contenido o música), aplicando filtros de privacidad y bloqueos
func (r *PostRepository) SearchPosts(ctx context.Context, searchTerm string, requesterID string, limit int, offset int) ([]Post, error) {
	searchPattern := "%" + searchTerm + "%"
	
	query := `
		SELECT p.id, p.author_id, p.content, p.music_id, p.is_private, p.created_at, p.updated_at
		FROM posts p
		WHERE (p.content ILIKE $1 OR p.music_id ILIKE $1)
		AND p.is_private = false
		AND p.author_id NOT IN (
			SELECT blocked_id FROM user_blocks WHERE blocker_id = $2
		)
		ORDER BY p.created_at DESC
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
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Content, &p.MusicID, &p.IsPrivate, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	
	if posts == nil {
		posts = []Post{}
	}
	return posts, nil
}