package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

// UpdatePost modifica el contenido de una publicación usando el esquema real del MER
func (r *PostRepository) UpdatePost(ctx context.Context, postID string, authorID string, newContent string) error {
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
func (r *PostRepository) DeletePost(ctx context.Context, postID string, authorID string) error {
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