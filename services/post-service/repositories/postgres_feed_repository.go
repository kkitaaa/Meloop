package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const maxFeedCandidates = 200
const postgresMaxParameters = 65535

type FeedRepository interface {
	ListCandidates(ctx context.Context, userID string, recommendedIDs, blockedAuthorIDs []string, limit int) ([]FeedCandidate, error)
	GetPostsByIDs(ctx context.Context, userID string, postIDs, blockedAuthorIDs []string) ([]Post, error)
}

type FeedCandidate struct {
	Post
	CommentsCount int
}

type PostgresFeedRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewPostgresFeedRepository(db *gorm.DB, timeouts ...time.Duration) *PostgresFeedRepository {
	timeout := 220 * time.Millisecond
	if len(timeouts) > 0 && timeouts[0] > 0 {
		timeout = timeouts[0]
	}
	return &PostgresFeedRepository{db: db, timeout: timeout}
}

func (r *PostgresFeedRepository) ListCandidates(ctx context.Context, userID string, recommendedIDs, blockedAuthorIDs []string, limit int) ([]FeedCandidate, error) {
	if len(recommendedIDs) > maxFeedCandidates {
		recommendedIDs = recommendedIDs[:maxFeedCandidates]
	}
	if limit < 1 || limit > maxFeedCandidates {
		limit = maxFeedCandidates
	}

	query := `
		SELECT p.id_publicacion AS id, p.id_usuario AS author_id, p.texto AS content,
			p.id_cancion AS music_id, p.url_multimedia AS url_multimedia, p.fecha_creacion AS created_at,
			COALESCE(l.likes_count, 0) AS likes_count,
			EXISTS (
				SELECT 1 FROM interaccion il
				WHERE il.id_publicacion = p.id_publicacion
					AND il.id_usuario = ?
					AND LOWER(il.tipo) = 'like'
			) AS is_liked,
			COALESCE(c.comments_count, 0) AS comments_count
		FROM publicacion p
		JOIN usuario u ON u.id_usuario = p.id_usuario
		LEFT JOIN configuracion_privacidad privacy ON privacy.id_usuario = p.id_usuario
		LEFT JOIN (
			SELECT id_publicacion, COUNT(*) AS likes_count
			FROM interaccion WHERE LOWER(tipo) = 'like' GROUP BY id_publicacion
		) l ON l.id_publicacion = p.id_publicacion
		LEFT JOIN (
			SELECT id_publicacion, COUNT(*) AS comments_count
			FROM interaccion WHERE LOWER(tipo) = 'comentario' GROUP BY id_publicacion
		) c ON c.id_publicacion = p.id_publicacion
		WHERE p.fecha_creacion >= NOW() - INTERVAL '7 days'
			AND (
				p.id_usuario = ?
				OR UPPER(TRIM(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO'))) = 'PUBLICO'
				OR (
					UPPER(TRIM(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO'))) = 'AMIGOS'
					AND EXISTS (
						SELECT 1 FROM amistad a
						WHERE UPPER(TRIM(a.estado)) = 'ACEPTADA'
							AND (
								(a.id_usuario_1 = ? AND a.id_usuario_2 = p.id_usuario)
								OR (a.id_usuario_2 = ? AND a.id_usuario_1 = p.id_usuario)
							)
					)
				)
				OR (%s)
			)
			AND COALESCE(u.suspendido, FALSE) = FALSE
			AND p.id_usuario::text <> ALL(?::text[])
			AND NOT EXISTS (
				SELECT 1 FROM accion_moderacion am
				WHERE am.id_usuario_afectado = p.id_usuario
					AND UPPER(TRIM(am.tipo_accion)) IN ('BAN', 'SUSPENSION', 'INACTIVO', 'DESACTIVADO')
			)
			AND (
				(u.fecha_creacion IS NOT NULL AND u.fecha_creacion >= DATE_TRUNC('day', NOW()) - INTERVAL '6 months')
				OR (u.fecha_ultimo_login IS NOT NULL AND u.fecha_ultimo_login >= DATE_TRUNC('day', NOW()) - INTERVAL '6 months')
				OR EXISTS (
					SELECT 1 FROM publicacion active_post
					WHERE active_post.id_usuario = p.id_usuario
						AND active_post.fecha_creacion >= DATE_TRUNC('day', NOW()) - INTERVAL '6 months'
				)
				OR EXISTS (
					SELECT 1 FROM interaccion active_interaction
					WHERE active_interaction.id_usuario = p.id_usuario
						AND active_interaction.fecha_creacion >= DATE_TRUNC('day', NOW()) - INTERVAL '6 months'
				)
				OR (
					u.fecha_creacion IS NULL
					AND u.fecha_ultimo_login IS NULL
					AND NOT EXISTS (SELECT 1 FROM publicacion old_post WHERE old_post.id_usuario = p.id_usuario)
					AND NOT EXISTS (SELECT 1 FROM interaccion old_interaction WHERE old_interaction.id_usuario = p.id_usuario)
				)
			)
		ORDER BY (%s) DESC, p.fecha_creacion DESC, p.id_publicacion
		LIMIT ?
	`

	recommendedClause := "FALSE"
	recommendedOrder := "FALSE"
	args := []any{userID, userID, userID, userID, postgresTextArray(blockedAuthorIDs)}
	if len(recommendedIDs) > 0 {
		idPlaceholders := strings.TrimSuffix(strings.Repeat("?,", len(recommendedIDs)), ",")
		recommendedClause = fmt.Sprintf(`
			p.id_publicacion IN (%s)
			AND (
				UPPER(TRIM(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO'))) = 'PUBLICO'
				OR (
					UPPER(TRIM(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO'))) = 'AMIGOS'
					AND EXISTS (
						SELECT 1 FROM amistad recommended_friend
						WHERE UPPER(TRIM(recommended_friend.estado)) = 'ACEPTADA'
							AND (
								(recommended_friend.id_usuario_1 = ? AND recommended_friend.id_usuario_2 = p.id_usuario)
								OR (recommended_friend.id_usuario_2 = ? AND recommended_friend.id_usuario_1 = p.id_usuario)
							)
					)
				)
			)
		`, idPlaceholders)
		args = append(args, stringArgsToAny(recommendedIDs)...)
		args = append(args, userID, userID)
		recommendedOrder = "p.id_publicacion IN (" + idPlaceholders + ")"
		args = append(args, stringArgsToAny(recommendedIDs)...)
	}
	query = fmt.Sprintf(query, recommendedClause, recommendedOrder)
	args = append(args, limit)

	if len(args) > postgresMaxParameters {
		return nil, fmt.Errorf("feed candidate query has %d parameters, maximum is %d", len(args), postgresMaxParameters)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var candidates []FeedCandidate
	if err := r.db.WithContext(queryCtx).Raw(query, args...).Scan(&candidates).Error; err != nil {
		return nil, fmt.Errorf("query feed candidates: %w", err)
	}
	if candidates == nil {
		candidates = []FeedCandidate{}
	}
	return candidates, nil
}

func (r *PostgresFeedRepository) GetPostsByIDs(ctx context.Context, userID string, postIDs, blockedAuthorIDs []string) ([]Post, error) {
	if len(postIDs) == 0 {
		return []Post{}, nil
	}
	if len(postIDs) > maxFeedCandidates {
		return nil, fmt.Errorf("feed detail query has %d IDs, maximum is %d", len(postIDs), maxFeedCandidates)
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(postIDs)), ",")
	query := fmt.Sprintf(`
		SELECT p.id_publicacion AS id, p.id_usuario AS author_id, p.texto AS content,
			p.id_cancion AS music_id, p.url_multimedia AS url_multimedia, p.fecha_creacion AS created_at,
			COALESCE(l.likes_count, 0) AS likes_count,
			EXISTS (
				SELECT 1 FROM interaccion i
				WHERE i.id_publicacion = p.id_publicacion
					AND i.id_usuario = ?
					AND LOWER(i.tipo) = 'like'
			) AS is_liked
		FROM publicacion p
		JOIN usuario u ON u.id_usuario = p.id_usuario
		LEFT JOIN configuracion_privacidad privacy ON privacy.id_usuario = p.id_usuario
		LEFT JOIN (
			SELECT id_publicacion, COUNT(*) AS likes_count
			FROM interaccion WHERE LOWER(tipo) = 'like' GROUP BY id_publicacion
		) l ON l.id_publicacion = p.id_publicacion
		WHERE p.id_publicacion IN (%s)
			AND (
				p.id_usuario = ?
				OR UPPER(TRIM(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO'))) = 'PUBLICO'
				OR (
					UPPER(TRIM(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO'))) = 'AMIGOS'
					AND EXISTS (
						SELECT 1 FROM amistad a
						WHERE UPPER(TRIM(a.estado)) = 'ACEPTADA'
							AND (
								(a.id_usuario_1 = ? AND a.id_usuario_2 = p.id_usuario)
								OR (a.id_usuario_2 = ? AND a.id_usuario_1 = p.id_usuario)
							)
					)
				)
			)
			AND COALESCE(u.suspendido, FALSE) = FALSE
			AND p.id_usuario::text <> ALL(?::text[])
			AND NOT EXISTS (
				SELECT 1 FROM accion_moderacion am
				WHERE am.id_usuario_afectado = p.id_usuario
					AND UPPER(TRIM(am.tipo_accion)) IN ('BAN', 'SUSPENSION', 'INACTIVO', 'DESACTIVADO')
			)
			AND (
				(u.fecha_creacion IS NOT NULL AND u.fecha_creacion >= DATE_TRUNC('day', NOW()) - INTERVAL '6 months')
				OR (u.fecha_ultimo_login IS NOT NULL AND u.fecha_ultimo_login >= DATE_TRUNC('day', NOW()) - INTERVAL '6 months')
				OR EXISTS (
					SELECT 1 FROM publicacion active_post
					WHERE active_post.id_usuario = p.id_usuario
						AND active_post.fecha_creacion >= DATE_TRUNC('day', NOW()) - INTERVAL '6 months'
				)
				OR EXISTS (
					SELECT 1 FROM interaccion active_interaction
					WHERE active_interaction.id_usuario = p.id_usuario
						AND active_interaction.fecha_creacion >= DATE_TRUNC('day', NOW()) - INTERVAL '6 months'
				)
				OR (
					u.fecha_creacion IS NULL
					AND u.fecha_ultimo_login IS NULL
					AND NOT EXISTS (SELECT 1 FROM publicacion old_post WHERE old_post.id_usuario = p.id_usuario)
					AND NOT EXISTS (SELECT 1 FROM interaccion old_interaction WHERE old_interaction.id_usuario = p.id_usuario)
				)
			)
	`, placeholders)

	args := make([]any, 0, len(postIDs)+4)
	args = append(args, userID)
	for _, postID := range postIDs {
		args = append(args, postID)
	}
	args = append(args, userID, userID, userID, postgresTextArray(blockedAuthorIDs))
	if len(args) > postgresMaxParameters {
		return nil, fmt.Errorf("feed detail query has %d parameters, maximum is %d", len(args), postgresMaxParameters)
	}

	var posts []Post
	queryCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := r.db.WithContext(queryCtx).Raw(query, args...).Scan(&posts).Error; err != nil {
		return nil, fmt.Errorf("query feed post details: %w", err)
	}
	if posts == nil {
		posts = []Post{}
	}
	return posts, nil
}

func stringArgsToAny(values []string) []any {
	args := make([]any, len(values))
	for index, value := range values {
		args[index] = value
	}
	return args
}

func uniqueStrings(values []string) []string {
	unique := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}
