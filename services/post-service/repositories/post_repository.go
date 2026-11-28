package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPostNotFoundOrUnauthorized = errors.New("publicación no encontrada o no tienes permisos de autor")
	ErrAlreadyLiked               = errors.New("el usuario ya dio me gusta a esta publicación")
	ErrPostNotFound               = errors.New("publicación no encontrada")
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
	return r.GetPostByIDExcluding(ctx, postID, requesterID, nil)
}

func (r *SQLPostRepository) GetPostByIDExcluding(ctx context.Context, postID, requesterID string, excludedAuthorIDs []string) (*Post, error) {
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
			AND p.id_usuario::text <> ALL($3::text[])
	`
	row := r.db.QueryRowContext(ctx, query, postID, requesterID, postgresTextArray(excludedAuthorIDs))

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
		if err == sql.ErrNoRows {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("error al obtener post: %w", err)
	}

	return &post, nil
}

// GetPostsByAuthor obtiene los posts de un perfil con paginación
func (r *SQLPostRepository) GetPostsByAuthor(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]Post, error) {
	return r.GetPostsByAuthorExcluding(ctx, authorID, requesterID, limit, offset, nil)
}

func (r *SQLPostRepository) GetPostsByAuthorExcluding(ctx context.Context, authorID, requesterID string, limit, offset int, excludedAuthorIDs []string) ([]Post, error) {
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
			AND p.id_usuario::text <> ALL($5::text[])
		ORDER BY p.fecha_creacion DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, authorID, requesterID, limit, offset, postgresTextArray(excludedAuthorIDs))
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
	return r.SearchPostsExcluding(ctx, searchTerm, requesterID, limit, offset, nil)
}

func (r *SQLPostRepository) SearchPostsExcluding(ctx context.Context, searchTerm, requesterID string, limit, offset int, excludedAuthorIDs []string) ([]Post, error) {
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
			AND p.id_usuario::text <> ALL($5::text[])
		ORDER BY p.fecha_creacion DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, searchPattern, requesterID, limit, offset, postgresTextArray(excludedAuthorIDs))
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

func (r *SQLPostRepository) ListUserIDs(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id_usuario FROM usuario`)
	if err != nil {
		return nil, fmt.Errorf("list user IDs: %w", err)
	}
	defer rows.Close()
	userIDs := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user IDs: %w", err)
	}
	return userIDs, nil
}

func (r *SQLPostRepository) GetPostAuthorID(ctx context.Context, postID string) (string, error) {
	var authorID string
	err := r.db.QueryRowContext(ctx, `SELECT id_usuario FROM publicacion WHERE id_publicacion = $1`, postID).Scan(&authorID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", sql.ErrNoRows
	}
	if err != nil {
		return "", fmt.Errorf("get post author: %w", err)
	}
	return authorID, nil
}

func (r *SQLPostRepository) SavePost(ctx context.Context, userID, postID string) error {
	_, err := r.db.ExecContext(ctx, `
		WITH save_request AS (
			SELECT $1::varchar AS interaction_id, $2::varchar AS user_id, $3::varchar AS post_id
		)
		INSERT INTO interaccion (id_interaccion, id_usuario, id_publicacion, tipo, fecha_creacion)
		SELECT save_request.interaction_id, save_request.user_id, save_request.post_id, 'save', NOW()
		FROM save_request
		WHERE NOT EXISTS (
			SELECT 1 FROM interaccion
			WHERE id_usuario = save_request.user_id
				AND id_publicacion = save_request.post_id
				AND LOWER(tipo) = 'save'
		)
	`, "save-"+uuid.NewString(), userID, postID)
	if err != nil {
		return fmt.Errorf("save post: %w", err)
	}
	return nil
}

func (r *SQLPostRepository) UnsavePost(ctx context.Context, userID, postID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM interaccion
		WHERE id_usuario = $1 AND id_publicacion = $2 AND LOWER(tipo) = 'save'
	`, userID, postID)
	if err != nil {
		return fmt.Errorf("unsave post: %w", err)
	}
	return nil
}

func (r *SQLPostRepository) GetSavedPosts(ctx context.Context, userID string, excludedAuthorIDs []string, limit, offset int) ([]Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id_publicacion, p.id_usuario, p.texto, p.id_cancion, p.url_multimedia, p.fecha_creacion,
			COALESCE(l.likes_count, 0),
			EXISTS (
				SELECT 1 FROM interaccion liked
				WHERE liked.id_publicacion = p.id_publicacion
					AND liked.id_usuario = $1
					AND LOWER(liked.tipo) = 'like'
			)
		FROM interaccion saved
		JOIN publicacion p ON p.id_publicacion = saved.id_publicacion
		LEFT JOIN (
			SELECT id_publicacion, COUNT(*) AS likes_count
			FROM interaccion
			WHERE LOWER(tipo) = 'like'
			GROUP BY id_publicacion
		) l ON l.id_publicacion = p.id_publicacion
		WHERE saved.id_usuario = $1
			AND LOWER(saved.tipo) = 'save'
			AND p.id_usuario::text <> ALL($2::text[])
		ORDER BY saved.fecha_creacion DESC
		LIMIT $3 OFFSET $4
	`, userID, postgresTextArray(excludedAuthorIDs), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get saved posts: %w", err)
	}
	defer rows.Close()
	posts := []Post{}
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Content, &post.MusicID, &post.UrlMultimedia, &post.CreatedAt, &post.LikesCount, &post.IsLiked); err != nil {
			return nil, fmt.Errorf("scan saved post: %w", err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate saved posts: %w", err)
	}
	return posts, nil
}

// CreatePost inserta una nueva publicación en la base de datos
func (r *SQLPostRepository) CreatePost(ctx context.Context, authorID string, content string, urlMultimedia *string) (*Post, error) {
	newID := uuid.NewString()

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
