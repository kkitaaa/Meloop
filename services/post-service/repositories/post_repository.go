package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/meloop/post-service/models"
)

var (
	ErrPostNotFound               = errors.New("publicación no encontrada")
	ErrUnauthorized               = errors.New("no tienes permisos para modificar esta publicación")
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
	LikesCount    int       `json:"likes_count"`
	IsLiked       bool      `json:"is_liked"`
}

// CreatePostParams agrupa los datos validados para insertar en una transacción
type CreatePostParams struct {
	AuthorID      string
	Content       string
	UrlMultimedia *string
	Music         *models.MusicReference
}

// PostRepository define el contrato de persistencia de publicaciones
type PostRepository interface {
	CreatePost(ctx context.Context, params CreatePostParams) (*Post, error)
	UpdatePost(ctx context.Context, postID string, authorID string, newContent string) error
	DeletePost(ctx context.Context, postID string, authorID string) (string, error)
	GetPostByID(ctx context.Context, postID string, requesterID string) (*Post, error)
	GetPostsByAuthor(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]Post, error)
	SearchPosts(ctx context.Context, searchTerm string, requesterID string, limit int, offset int) ([]Post, error)
	AddLike(ctx context.Context, postID string, userID string) error
	RemoveLike(ctx context.Context, postID string, userID string) error
}

// SQLPostRepository maneja la persistencia de las publicaciones
type SQLPostRepository struct {
	db *sql.DB
}

// NewPostRepository crea una nueva instancia del repositorio
func NewPostRepository(db *sql.DB) *SQLPostRepository {
	return &SQLPostRepository{db: db}
}

// CreatePost inserta la publicación (y su referencia en cancion si aplica) dentro de una transacción SQL (RF-16)
func (r *SQLPostRepository) CreatePost(ctx context.Context, params CreatePostParams) (*Post, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	var resolvedCancionID *string
	if params.Music != nil && strings.TrimSpace(params.Music.ReferenceID) != "" {
		refID := strings.TrimSpace(params.Music.ReferenceID)
		var existingID string

		findSongQuery := `SELECT id_cancion FROM cancion WHERE id_cancion = $1 OR spotify_id = $1 LIMIT 1`
		err := tx.QueryRowContext(ctx, findSongQuery, refID).Scan(&existingID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("error consultando canción: %w", err)
			}

			newCancionID := uuid.NewString()
			titulo := strings.TrimSpace(params.Music.Title)
			if titulo == "" {
				titulo = refID
			}
			artista := "Artista Meloop"

			insertSongQuery := `
				INSERT INTO cancion (id_cancion, spotify_id, titulo, artista_nombre)
				VALUES ($1, $2, $3, $4)
				RETURNING id_cancion
			`
			if err := tx.QueryRowContext(ctx, insertSongQuery, newCancionID, refID, titulo, artista).Scan(&existingID); err != nil {
				return nil, fmt.Errorf("error registrando referencia musical: %w", err)
			}
		}
		resolvedCancionID = &existingID
	}

	newPostID := uuid.NewString()
	insertPostQuery := `
		INSERT INTO publicacion (id_publicacion, id_usuario, id_cancion, texto, url_multimedia, fecha_creacion)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id_publicacion, id_usuario, id_cancion, texto, url_multimedia, fecha_creacion
	`

	var post Post
	err = tx.QueryRowContext(
		ctx,
		insertPostQuery,
		newPostID,
		params.AuthorID,
		resolvedCancionID,
		params.Content,
		params.UrlMultimedia,
	).Scan(
		&post.ID,
		&post.AuthorID,
		&post.MusicID,
		&post.Content,
		&post.UrlMultimedia,
		&post.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error insertando publicación: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error confirmando transacción: %w", err)
	}
	tx = nil

	return &post, nil
}

// classifyOwnershipFailure determina si 0 filas afectadas se debe a post inexistente (404) o sin permisos (403)
func (r *SQLPostRepository) classifyOwnershipFailure(ctx context.Context, postID string) error {
	var ownerID string
	query := `SELECT id_usuario FROM publicacion WHERE id_publicacion = $1`

	err := r.db.QueryRowContext(ctx, query, postID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPostNotFound
		}
		return fmt.Errorf("error al verificar publicación: %w", err)
	}

	return ErrUnauthorized
}

// UpdatePost modifica el contenido de una publicación de forma atómica (RF-17)
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
		return r.classifyOwnershipFailure(ctx, postID)
	}

	return nil
}

// DeletePost elimina atómicamente una publicación y retorna su url_multimedia para limpieza en MinIO (RF-18)
func (r *SQLPostRepository) DeletePost(ctx context.Context, postID string, authorID string) (string, error) {
	query := `
		DELETE FROM publicacion 
		WHERE id_publicacion = $1 AND id_usuario = $2
		RETURNING COALESCE(url_multimedia, '')
	`

	var mediaURL string
	err := r.db.QueryRowContext(ctx, query, postID, authorID).Scan(&mediaURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", r.classifyOwnershipFailure(ctx, postID)
		}
		return "", fmt.Errorf("error al eliminar la publicación: %w", err)
	}

	return mediaURL, nil
}

// GetPostByID obtiene un post por su ID
func (r *SQLPostRepository) GetPostByID(ctx context.Context, postID string, requesterID string) (*Post, error) {
	query := `
		SELECT p.id_publicacion, p.id_usuario, p.texto, p.id_cancion, p.url_multimedia, p.fecha_creacion,
			COALESCE(l.likes_count, 0),
			EXISTS (
				SELECT 1 FROM interaccion i
				WHERE i.id_publicacion = p.id_publicacion
					AND i.id_usuario = $2
					AND i.tipo = 'like'
			)
		FROM publicacion p
		LEFT JOIN (
			SELECT id_publicacion, COUNT(id_interaccion) AS likes_count
			FROM interaccion
			WHERE tipo = 'like'
			GROUP BY id_publicacion
		) l ON p.id_publicacion = l.id_publicacion
		WHERE p.id_publicacion = $1
	`
	row := r.db.QueryRowContext(ctx, query, postID, requesterID)

	var post Post
	err := row.Scan(
		&post.ID,
		&post.AuthorID,
		&post.Content,
		&post.MusicID,
		&post.UrlMultimedia,
		&post.CreatedAt,
		&post.LikesCount,
		&post.IsLiked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("error al obtener post: %w", err)
	}

	return &post, nil
}

// GetPostsByAuthor obtiene los posts de un perfil con paginación
func (r *SQLPostRepository) GetPostsByAuthor(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]Post, error) {
	query := `
		SELECT p.id_publicacion, p.id_usuario, p.texto, p.id_cancion, p.url_multimedia, p.fecha_creacion,
			COALESCE(l.likes_count, 0),
			EXISTS (
				SELECT 1 FROM interaccion i
				WHERE i.id_publicacion = p.id_publicacion
					AND i.id_usuario = $2
					AND i.tipo = 'like'
			)
		FROM publicacion p
		LEFT JOIN (
			SELECT id_publicacion, COUNT(id_interaccion) AS likes_count
			FROM interaccion
			WHERE tipo = 'like'
			GROUP BY id_publicacion
		) l ON p.id_publicacion = l.id_publicacion
		WHERE p.id_usuario = $1
		ORDER BY p.fecha_creacion DESC
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
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Content,
			&p.MusicID,
			&p.UrlMultimedia,
			&p.CreatedAt,
			&p.LikesCount,
			&p.IsLiked,
		); err != nil {
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
		SELECT p.id_publicacion, p.id_usuario, p.texto, p.id_cancion, p.url_multimedia, p.fecha_creacion,
			COALESCE(l.likes_count, 0),
			EXISTS (
				SELECT 1 FROM interaccion i
				WHERE i.id_publicacion = p.id_publicacion
					AND i.id_usuario = $2
					AND i.tipo = 'like'
			)
		FROM publicacion p
		LEFT JOIN (
			SELECT id_publicacion, COUNT(id_interaccion) AS likes_count
			FROM interaccion
			WHERE tipo = 'like'
			GROUP BY id_publicacion
		) l ON p.id_publicacion = l.id_publicacion
		WHERE (p.texto ILIKE $1 OR p.id_cancion ILIKE $1)
		ORDER BY p.fecha_creacion DESC
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
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Content,
			&p.MusicID,
			&p.UrlMultimedia,
			&p.CreatedAt,
			&p.LikesCount,
			&p.IsLiked,
		); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	if posts == nil {
		posts = []Post{}
	}
	return posts, nil
}

// AddLike inserta una interacción de tipo 'like' validando que no exista previamente (RF-20)
func (r *SQLPostRepository) AddLike(ctx context.Context, postID string, userID string) error {
	checkQuery := `SELECT COUNT(1) FROM interaccion WHERE id_usuario = $1 AND id_publicacion = $2 AND tipo = 'like'`
	var count int
	if err := r.db.QueryRowContext(ctx, checkQuery, userID, postID).Scan(&count); err != nil {
		return fmt.Errorf("error al verificar el estado del like: %w", err)
	}

	if count > 0 {
		return ErrAlreadyLiked
	}

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
