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
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// UpdatePost modifica el contenido de una publicación solo si el autor coincide
func (r *PostRepository) UpdatePost(ctx context.Context, postID string, authorID string, newContent string) error {
	// La condición "AND author_id = $2" es el mecanismo de seguridad principal
	query := `
		UPDATE posts 
		SET content = $1, updated_at = $2 
		WHERE id = $3 AND author_id = $4
	`
	
	result, err := r.db.ExecContext(ctx, query, newContent, time.Now(), postID, authorID)
	if err != nil {
		return fmt.Errorf("error al actualizar la publicación: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	// Si no se afectó ninguna fila, el post no existe o el usuario no es el autor
	if rowsAffected == 0 {
		return ErrPostNotFoundOrUnauthorized
	}

	return nil
}

// DeletePost elimina una publicación validando la autoría
func (r *PostRepository) DeletePost(ctx context.Context, postID string, authorID string) error {
	query := `
		DELETE FROM posts 
		WHERE id = $1 AND author_id = $2
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