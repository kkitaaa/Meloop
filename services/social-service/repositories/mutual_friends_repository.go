package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/social-service/models"
)

type MutualFriendsReader interface {
	ListMutualFriends(ctx context.Context, viewerID, targetID, afterID string, limit int) (models.MutualFriendsData, error)
}

type postgresMutualFriendsRepository struct {
	pool *pgxpool.Pool
}

func NewMutualFriendsRepository(pool *pgxpool.Pool) MutualFriendsReader {
	return &postgresMutualFriendsRepository{pool: pool}
}

func (r *postgresMutualFriendsRepository) ListMutualFriends(
	ctx context.Context,
	viewerID, targetID, afterID string,
	limit int,
) (models.MutualFriendsData, error) {
	if r.pool == nil {
		return models.MutualFriendsData{}, errors.New("database connection pool is not initialized")
	}

	var exists bool
	var visibility string
	var blocked bool
	err := r.pool.QueryRow(ctx, `
		SELECT u.id_usuario IS NOT NULL,
		       COALESCE(cp.visibilidad_amigos, 'PUBLICO'),
		       EXISTS (
		           SELECT 1 FROM bloqueo b
		           WHERE (b.id_usuario_bloqueador = $1 AND b.id_usuario_bloqueado = $2)
		              OR (b.id_usuario_bloqueador = $2 AND b.id_usuario_bloqueado = $1)
		       )
		FROM (SELECT $2::varchar AS id_usuario) requested
		LEFT JOIN usuario u ON u.id_usuario = requested.id_usuario
		LEFT JOIN configuracion_privacidad cp ON cp.id_usuario = u.id_usuario`, viewerID, targetID).
		Scan(&exists, &visibility, &blocked)
	if err != nil {
		return models.MutualFriendsData{}, err
	}
	if !exists || blocked {
		return models.MutualFriendsData{}, ErrUserNotFound
	}

	var totalCount int
	err = r.pool.QueryRow(ctx, mutualFriendsCountQuery, viewerID, targetID).Scan(&totalCount)
	if err != nil {
		return models.MutualFriendsData{}, err
	}

	data := models.MutualFriendsData{
		Items:      make([]models.MutualFriend, 0),
		TotalCount: totalCount,
		Hidden:     strings.EqualFold(visibility, "OCULTO"),
	}
	if data.Hidden || totalCount == 0 {
		return data, nil
	}

	rows, err := r.pool.Query(ctx, mutualFriendsListQuery, viewerID, targetID, afterID, limit)
	if err != nil {
		return models.MutualFriendsData{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var friend models.MutualFriend
		if err := rows.Scan(&friend.ID, &friend.Username, &friend.AvatarURL); err != nil {
			return models.MutualFriendsData{}, err
		}
		data.Items = append(data.Items, friend)
	}
	if err := rows.Err(); err != nil {
		return models.MutualFriendsData{}, err
	}

	return data, nil
}

const mutualFriendsCTE = `
	WITH viewer_friends AS (
		SELECT CASE WHEN a.id_usuario_1::text = $1 THEN a.id_usuario_2::text ELSE a.id_usuario_1::text END AS friend_id
		FROM amistad a
		WHERE (a.id_usuario_1::text = $1 OR a.id_usuario_2::text = $1)
		  AND a.estado = 'ACEPTADA'
	),
	target_friends AS (
		SELECT CASE WHEN a.id_usuario_1::text = $2 THEN a.id_usuario_2::text ELSE a.id_usuario_1::text END AS friend_id
		FROM amistad a
		WHERE (a.id_usuario_1::text = $2 OR a.id_usuario_2::text = $2)
		  AND a.estado = 'ACEPTADA'
	),
	mutual_friends AS (
		SELECT DISTINCT vf.friend_id AS id_usuario
		FROM viewer_friends vf
		INNER JOIN target_friends tf ON tf.friend_id = vf.friend_id
		WHERE vf.friend_id NOT IN ($1, $2)
		  AND NOT EXISTS (
			SELECT 1 FROM bloqueo b
			WHERE (b.id_usuario_bloqueador::text = $1 AND b.id_usuario_bloqueado::text = vf.friend_id)
			   OR (b.id_usuario_bloqueador::text = vf.friend_id AND b.id_usuario_bloqueado::text = $1)
			   OR (b.id_usuario_bloqueador::text = $2 AND b.id_usuario_bloqueado::text = vf.friend_id)
			   OR (b.id_usuario_bloqueador::text = vf.friend_id AND b.id_usuario_bloqueado::text = $2)
		  )
	)
`

const mutualFriendsCountQuery = mutualFriendsCTE + `
	SELECT COUNT(*) FROM mutual_friends
`

const mutualFriendsListQuery = mutualFriendsCTE + `
	SELECT u.id_usuario::text, u.username, COALESCE(p.foto_perfil, '')
	FROM mutual_friends mf
	INNER JOIN usuario u ON u.id_usuario::text = mf.id_usuario
	LEFT JOIN perfil p ON p.id_usuario = u.id_usuario
	WHERE mf.id_usuario > $3
	ORDER BY mf.id_usuario
	LIMIT $4
`
