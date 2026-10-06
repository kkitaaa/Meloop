package repositories

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/recommendation-service/models"
)

const defaultActivityWindow = 30 * 24 * time.Hour

type UserProfileRepository interface {
	Get(ctx context.Context, userID, recommendationType string) (models.UserProfile, []models.Interaction, error)
}

type PopularityRepository interface {
	GetPopular(ctx context.Context, userID string, limit int) ([]models.PopularContent, error)
}

type FriendRecommendationEligibilityRepository interface {
	FilterEligibleFriendCandidates(ctx context.Context, userID string, candidateIDs []int) (map[int]struct{}, error)
}

type postgresUserProfileRepository struct{ pool *pgxpool.Pool }

func NewUserProfileRepository(pool *pgxpool.Pool) UserProfileRepository {
	return &postgresUserProfileRepository{pool: pool}
}

func (repository *postgresUserProfileRepository) Get(ctx context.Context, userID, recommendationType string) (models.UserProfile, []models.Interaction, error) {
	if repository.pool == nil {
		return models.UserProfile{}, nil, errors.New("database connection pool is not initialized")
	}
	var profile models.UserProfile
	rows, err := repository.pool.Query(ctx, `
		SELECT UPPER(TRIM(tipo)), TRIM(spotify_id)
		FROM PREFERENCIA_MUSICAL
		WHERE id_usuario = $1
		  AND spotify_id IS NOT NULL
		  AND tipo IS NOT NULL
		  AND TRIM(spotify_id) <> ''
		ORDER BY tipo, spotify_id`, userID)
	if err != nil {
		return models.UserProfile{}, nil, fmt.Errorf("query user preferences: %w", err)
	}
	for rows.Next() {
		var preferenceType, value string
		if err := rows.Scan(&preferenceType, &value); err != nil {
			rows.Close()
			return models.UserProfile{}, nil, fmt.Errorf("scan user preference: %w", err)
		}
		switch preferenceType {
		case "GENERO", "GENRE":
			profile.Genres = append(profile.Genres, value)
		case "ARTISTA", "ARTIST":
			profile.Artists = append(profile.Artists, value)
		case "CANCION", "SONG", "TRACK":
			profile.Songs = append(profile.Songs, value)
		}
	}
	if err := rows.Err(); err != nil {
		return models.UserProfile{}, nil, fmt.Errorf("read user preferences: %w", err)
	}
	rows.Close()

	err = repository.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM INTERACCION i
		WHERE i.id_usuario = $1
		  AND i.fecha_creacion >= now() - ($2 * interval '1 second')
		  AND NOT EXISTS (
			SELECT 1 FROM CONFIGURACION_PRIVACIDAD privacy
			WHERE privacy.id_usuario = $1
			  AND UPPER(TRIM(COALESCE(privacy.visibilidad_interacciones, 'PUBLICO'))) = 'PRIVADO'
		  )`, userID, int64(defaultActivityWindow/time.Second)).Scan(&profile.InteractionCount)
	if err != nil {
		return models.UserProfile{}, nil, fmt.Errorf("count recent user interactions: %w", err)
	}

	rows, err = repository.pool.Query(ctx, `
		SELECT COALESCE(LOWER(TRIM(i.tipo)), ''), i.id_publicacion
		FROM INTERACCION i
		WHERE i.id_usuario = $1
		  AND i.fecha_creacion >= now() - ($2 * interval '1 second')
		  AND NOT EXISTS (
			SELECT 1 FROM CONFIGURACION_PRIVACIDAD privacy
			WHERE privacy.id_usuario = $1
			  AND UPPER(TRIM(COALESCE(privacy.visibilidad_interacciones, 'PUBLICO'))) = 'PRIVADO'
		  )
		ORDER BY i.fecha_creacion DESC
		LIMIT 100`, userID, int64(defaultActivityWindow/time.Second))
	if err != nil {
		return models.UserProfile{}, nil, fmt.Errorf("query user interactions: %w", err)
	}
	defer rows.Close()
	var interactions []models.Interaction
	for rows.Next() {
		var interaction models.Interaction
		var targetID string
		if err := rows.Scan(&interaction.Type, &targetID); err != nil {
			return models.UserProfile{}, nil, fmt.Errorf("scan recent user interaction: %w", err)
		}
		interaction.TargetID, err = strconv.Atoi(targetID)
		if err != nil || interaction.TargetID < 1 {
			continue
		}
		if recommendationType == "music" && interaction.Type != "like" {
			continue
		}
		if recommendationType == "friends" && interaction.Type != "friend_added" {
			continue
		}
		interactions = append(interactions, interaction)
	}
	if err := rows.Err(); err != nil {
		return models.UserProfile{}, nil, fmt.Errorf("read user interactions: %w", err)
	}
	return profile, interactions, nil
}

func (repository *postgresUserProfileRepository) GetPopular(ctx context.Context, userID string, limit int) ([]models.PopularContent, error) {
	if repository.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	if limit < 1 {
		return []models.PopularContent{}, nil
	}

	rows, err := repository.pool.Query(ctx, `
		WITH active_users AS (
			SELECT u.id_usuario
			FROM USUARIO u
			WHERE NOT EXISTS (
				SELECT 1
				FROM ACCION_MODERACION am
				WHERE am.id_usuario_afectado = u.id_usuario
				  AND UPPER(am.tipo_accion) IN ('BAN', 'SUSPENSION', 'INACTIVO', 'DESACTIVADO')
			)
		),
		popular_content AS (
			SELECT 'song'::text AS item_type,
			       COALESCE(NULLIF(TRIM(c.spotify_id), ''), c.id_cancion::text) AS item_id,
			       COALESCE(NULLIF(TRIM(c.titulo), ''), c.id_cancion::text) AS item_name
			FROM PUBLICACION p
			JOIN active_users author ON author.id_usuario = p.id_usuario
			JOIN CANCION c ON c.id_cancion = p.id_cancion
			LEFT JOIN CONFIGURACION_PRIVACIDAD privacy ON privacy.id_usuario = p.id_usuario
			WHERE p.fecha_creacion >= now() - ($2 * interval '1 second')
			  AND UPPER(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO')) = 'PUBLICO'
			  AND NOT EXISTS (
				SELECT 1 FROM BLOQUEO b
				WHERE (b.id_usuario_bloqueador::text = $1 AND b.id_usuario_bloqueado = p.id_usuario)
				   OR (b.id_usuario_bloqueado::text = $1 AND b.id_usuario_bloqueador = p.id_usuario)
			  )
			UNION ALL
			SELECT 'artist'::text,
			       TRIM(c.artista_nombre),
			       TRIM(c.artista_nombre)
			FROM PUBLICACION p
			JOIN active_users author ON author.id_usuario = p.id_usuario
			JOIN CANCION c ON c.id_cancion = p.id_cancion
			LEFT JOIN CONFIGURACION_PRIVACIDAD privacy ON privacy.id_usuario = p.id_usuario
			WHERE p.fecha_creacion >= now() - ($2 * interval '1 second')
			  AND NULLIF(TRIM(c.artista_nombre), '') IS NOT NULL
			  AND UPPER(COALESCE(privacy.visibilidad_publicaciones, 'PUBLICO')) = 'PUBLICO'
			  AND NOT EXISTS (
				SELECT 1 FROM BLOQUEO b
				WHERE (b.id_usuario_bloqueador::text = $1 AND b.id_usuario_bloqueado = p.id_usuario)
				   OR (b.id_usuario_bloqueado::text = $1 AND b.id_usuario_bloqueador = p.id_usuario)
			  )
			UNION ALL
			SELECT CASE
					WHEN UPPER(TRIM(pref.tipo)) IN ('ARTISTA', 'ARTIST') THEN 'artist'
					ELSE 'song'
			       END,
			       TRIM(pref.spotify_id),
			       TRIM(pref.spotify_id)
			FROM PREFERENCIA_MUSICAL pref
			JOIN active_users author ON author.id_usuario = pref.id_usuario
			WHERE pref.fecha_creacion >= now() - ($2 * interval '1 second')
			  AND pref.spotify_id IS NOT NULL
			  AND TRIM(pref.spotify_id) <> ''
			  AND UPPER(TRIM(pref.tipo)) IN ('ARTISTA', 'ARTIST', 'CANCION', 'SONG', 'TRACK')
			  AND NOT EXISTS (
				SELECT 1 FROM BLOQUEO b
				WHERE (b.id_usuario_bloqueador::text = $1 AND b.id_usuario_bloqueado = pref.id_usuario)
				   OR (b.id_usuario_bloqueado::text = $1 AND b.id_usuario_bloqueador = pref.id_usuario)
			  )
		)
		SELECT item_type, item_id, MAX(item_name), COUNT(*) AS usage_count
		FROM popular_content
		WHERE item_id IS NOT NULL AND TRIM(item_id) <> ''
		GROUP BY item_type, item_id
		ORDER BY usage_count DESC, item_type, item_id
		LIMIT $3`, userID, int64(defaultActivityWindow/time.Second), limit)
	if err != nil {
		return nil, fmt.Errorf("query popular recommendation content: %w", err)
	}
	defer rows.Close()

	popular := make([]models.PopularContent, 0, limit)
	for rows.Next() {
		var item models.PopularContent
		if err := rows.Scan(&item.Type, &item.ID, &item.Name, &item.UsageCount); err != nil {
			return nil, fmt.Errorf("scan popular recommendation content: %w", err)
		}
		popular = append(popular, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read popular recommendation content: %w", err)
	}
	return popular, nil
}

func (repository *postgresUserProfileRepository) FilterEligibleFriendCandidates(ctx context.Context, userID string, candidateIDs []int) (map[int]struct{}, error) {
	if repository.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	eligible := make(map[int]struct{}, len(candidateIDs))
	if len(candidateIDs) == 0 {
		return eligible, nil
	}

	candidateUserIDs := make([]string, 0, len(candidateIDs))
	for _, candidateID := range candidateIDs {
		if candidateID > 0 {
			candidateUserIDs = append(candidateUserIDs, strconv.Itoa(candidateID))
		}
	}
	if len(candidateUserIDs) == 0 {
		return eligible, nil
	}

	rows, err := repository.pool.Query(ctx, `
		SELECT DISTINCT u.id_usuario::text
		FROM UNNEST($2::text[]) AS candidates(candidate_id)
		JOIN USUARIO u ON u.id_usuario::text = candidates.candidate_id
		WHERE u.id_usuario::text <> $1
		  AND NOT EXISTS (
			SELECT 1
			FROM BLOQUEO b
			WHERE (b.id_usuario_bloqueador::text = $1 AND b.id_usuario_bloqueado::text = u.id_usuario::text)
			   OR (b.id_usuario_bloqueador::text = u.id_usuario::text AND b.id_usuario_bloqueado::text = $1)
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM ACCION_MODERACION am
			WHERE am.id_usuario_afectado::text = u.id_usuario::text
			  AND UPPER(TRIM(COALESCE(am.tipo_accion, ''))) IN ('BAN', 'SUSPENSION', 'INACTIVO', 'DESACTIVADO')
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM AMISTAD a
			WHERE ((a.id_usuario_1::text = $1 AND a.id_usuario_2::text = u.id_usuario::text)
			    OR (a.id_usuario_1::text = u.id_usuario::text AND a.id_usuario_2::text = $1))
			  AND UPPER(TRIM(COALESCE(a.estado, ''))) = 'ACEPTADA'
		  )`, userID, candidateUserIDs)
	if err != nil {
		return nil, fmt.Errorf("filter ineligible friend recommendation candidates: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var candidateID string
		if err := rows.Scan(&candidateID); err != nil {
			return nil, fmt.Errorf("scan eligible friend recommendation candidate: %w", err)
		}
		id, err := strconv.Atoi(candidateID)
		if err != nil {
			return nil, fmt.Errorf("parse eligible friend recommendation candidate id %q: %w", candidateID, err)
		}
		eligible[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read eligible friend recommendation candidates: %w", err)
	}
	return eligible, nil
}
