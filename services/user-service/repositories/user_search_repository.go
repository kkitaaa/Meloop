package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/models"
)

type UserSearchRepository interface {
	SearchPublicUsers(ctx context.Context, requesterID, query string, limit, offset int) ([]models.UserSearchResult, error)
}

type postgresUserSearchRepository struct {
	pool *pgxpool.Pool
}

func NewUserSearchRepository(pool *pgxpool.Pool) UserSearchRepository {
	return &postgresUserSearchRepository{pool: pool}
}

func (r *postgresUserSearchRepository) SearchPublicUsers(
	ctx context.Context,
	requesterID, query string,
	limit, offset int,
) ([]models.UserSearchResult, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	escapedQuery := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(query)
	pattern := "%" + escapedQuery + "%"
	rows, err := r.pool.Query(ctx, `
		SELECT
			u.id_usuario::text,
			u.username,
			COALESCE(p.foto_perfil, ''),
			COALESCE(p.biografia, '')
		FROM USUARIO u
		LEFT JOIN PERFIL p ON p.id_usuario = u.id_usuario
		LEFT JOIN CONFIGURACION_PRIVACIDAD cp ON cp.id_usuario = u.id_usuario
		WHERE u.id_usuario <> $2
		  AND COALESCE(u.suspendido, FALSE) = FALSE
		  AND COALESCE(UPPER(BTRIM(cp.visibilidad_perfil)), 'PUBLICO') = 'PUBLICO'
		  AND NOT EXISTS (
			  SELECT 1
			  FROM BLOQUEO b
			  WHERE (b.id_usuario_bloqueador = u.id_usuario AND b.id_usuario_bloqueado = $2)
			     OR (b.id_usuario_bloqueador = $2 AND b.id_usuario_bloqueado = u.id_usuario)
		  )
		  AND (
			  u.username ILIKE $3 ESCAPE '\'
			  OR p.biografia ILIKE $3 ESCAPE '\'
			  OR u.username % $1
			  OR p.biografia % $1
		  )
		ORDER BY
			CASE
				WHEN LOWER(u.username) = LOWER($1) THEN 0
				WHEN LOWER(u.username) LIKE LOWER($1) || '%' THEN 1
				ELSE 2
			END,
			GREATEST(
				similarity(u.username, $1),
				similarity(p.biografia, $1)
			) DESC,
			LOWER(u.username),
			u.id_usuario
		LIMIT $4 OFFSET $5
	`, query, requesterID, pattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("buscar usuarios públicos: %w", err)
	}
	defer rows.Close()

	results := make([]models.UserSearchResult, 0)
	for rows.Next() {
		var result models.UserSearchResult
		if err := rows.Scan(&result.IDUsuario, &result.Username, &result.FotoPerfil, &result.Biografia); err != nil {
			return nil, fmt.Errorf("leer resultado de búsqueda de usuarios: %w", err)
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar resultados de búsqueda de usuarios: %w", err)
	}
	return results, nil
}
