package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/models"
)

// UserRepository define las operaciones de acceso a datos para la entidad USUARIO
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.Usuario, error)
	GetByUsername(ctx context.Context, username string) (*models.Usuario, error)
	GetByEmail(ctx context.Context, email string) (*models.Usuario, error)
	UpdateUsername(ctx context.Context, id string, newUsername string) (*models.Usuario, error)
	ListUsers(ctx context.Context) ([]models.AdminUserSummary, error)
	SuspendUser(ctx context.Context, id string, reason string) (*models.AdminUserSummary, error)
	ReactivateUser(ctx context.Context, id string) (*models.AdminUserSummary, error)
	SetModeratorRole(ctx context.Context, id string, moderator bool) (*models.AdminUserSummary, error)
}

// postgresUserRepository implementa UserRepository utilizando un pool de PostgreSQL
type postgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository crea una nueva instancia del repositorio de usuarios
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{pool: pool}
}

// GetByID busca y devuelve un usuario por su identificador único
func (r *postgresUserRepository) GetByID(ctx context.Context, id string) (*models.Usuario, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	var user models.Usuario
	query := `
		SELECT id_usuario::text, username, correo, contrasena_hash, id_nivel, experiencia
		FROM USUARIO
		WHERE id_usuario = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.IDUsuario,
		&user.Username,
		&user.Correo,
		&user.ContrasenaHash,
		&user.IDNivel,
		&user.Experiencia,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername busca un usuario por su nombre de usuario
func (r *postgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.Usuario, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	var user models.Usuario
	query := `
		SELECT id_usuario::text, username, correo, contrasena_hash, id_nivel, experiencia
		FROM USUARIO
		WHERE username = $1
	`
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.IDUsuario,
		&user.Username,
		&user.Correo,
		&user.ContrasenaHash,
		&user.IDNivel,
		&user.Experiencia,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail busca un usuario por su correo electrónico
func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.Usuario, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	var user models.Usuario
	query := `
		SELECT id_usuario::text, username, correo, contrasena_hash, id_nivel, experiencia
		FROM USUARIO
		WHERE correo = $1
	`
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.IDUsuario,
		&user.Username,
		&user.Correo,
		&user.ContrasenaHash,
		&user.IDNivel,
		&user.Experiencia,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUsername actualiza el nombre de usuario de un usuario existente
func (r *postgresUserRepository) UpdateUsername(ctx context.Context, id string, newUsername string) (*models.Usuario, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	var user models.Usuario
	query := `
		UPDATE USUARIO
		SET username = $1
		WHERE id_usuario = $2
		RETURNING id_usuario::text, username, correo, contrasena_hash, id_nivel, experiencia
	`
	err := r.pool.QueryRow(ctx, query, newUsername, id).Scan(
		&user.IDUsuario,
		&user.Username,
		&user.Correo,
		&user.ContrasenaHash,
		&user.IDNivel,
		&user.Experiencia,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) ensureAdminColumns(ctx context.Context) error {
	if r.pool == nil {
		return errors.New("el pool de base de datos no está inicializado")
	}
	queries := []string{
		`ALTER TABLE USUARIO ADD COLUMN IF NOT EXISTS rol VARCHAR(30) NOT NULL DEFAULT 'usuario'`,
		`ALTER TABLE USUARIO ADD COLUMN IF NOT EXISTS suspendido BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE USUARIO ADD COLUMN IF NOT EXISTS es_moderador BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE USUARIO ADD COLUMN IF NOT EXISTS motivo_suspension TEXT`,
	}
	for _, q := range queries {
		if _, err := r.pool.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (r *postgresUserRepository) ListUsers(ctx context.Context) ([]models.AdminUserSummary, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}
	if err := r.ensureAdminColumns(ctx); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id_usuario::text, username, correo,
			COALESCE(rol, 'usuario') AS rol,
			COALESCE(suspendido, FALSE) AS suspendido,
			COALESCE(es_moderador, FALSE) AS es_moderador,
			motivo_suspension
		FROM USUARIO
		ORDER BY username ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.AdminUserSummary, 0)
	for rows.Next() {
		var user models.AdminUserSummary
		if err := rows.Scan(&user.IDUsuario, &user.Username, &user.Correo, &user.Rol, &user.Suspendido, &user.EsModerador, &user.Motivo); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *postgresUserRepository) SuspendUser(ctx context.Context, id string, reason string) (*models.AdminUserSummary, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}
	if err := r.ensureAdminColumns(ctx); err != nil {
		return nil, err
	}

	var user models.AdminUserSummary
	query := `
		UPDATE USUARIO
		SET suspendido = TRUE,
		    motivo_suspension = $1,
		    rol = CASE WHEN rol = 'admin' THEN 'admin' ELSE 'usuario' END
		WHERE id_usuario = $2
		RETURNING id_usuario::text, username, correo,
			COALESCE(rol, 'usuario') AS rol,
			COALESCE(suspendido, FALSE) AS suspendido,
			COALESCE(es_moderador, FALSE) AS es_moderador,
			motivo_suspension
	`
	err := r.pool.QueryRow(ctx, query, reason, id).Scan(&user.IDUsuario, &user.Username, &user.Correo, &user.Rol, &user.Suspendido, &user.EsModerador, &user.Motivo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) ReactivateUser(ctx context.Context, id string) (*models.AdminUserSummary, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}
	if err := r.ensureAdminColumns(ctx); err != nil {
		return nil, err
	}

	var user models.AdminUserSummary
	query := `
		UPDATE USUARIO
		SET suspendido = FALSE,
		    motivo_suspension = NULL
		WHERE id_usuario = $1
		RETURNING id_usuario::text, username, correo,
			COALESCE(rol, 'usuario') AS rol,
			COALESCE(suspendido, FALSE) AS suspendido,
			COALESCE(es_moderador, FALSE) AS es_moderador,
			motivo_suspension
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.IDUsuario, &user.Username, &user.Correo, &user.Rol, &user.Suspendido, &user.EsModerador, &user.Motivo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) SetModeratorRole(ctx context.Context, id string, moderator bool) (*models.AdminUserSummary, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}
	if err := r.ensureAdminColumns(ctx); err != nil {
		return nil, err
	}

	var user models.AdminUserSummary
	query := `
		UPDATE USUARIO
		SET es_moderador = $1,
		    rol = CASE
				WHEN $1 THEN CASE WHEN rol = 'admin' THEN 'admin' ELSE 'moderador' END
				ELSE CASE WHEN rol = 'admin' THEN 'admin' ELSE 'usuario' END
			END
		WHERE id_usuario = $2
		RETURNING id_usuario::text, username, correo,
			COALESCE(rol, 'usuario') AS rol,
			COALESCE(suspendido, FALSE) AS suspendido,
			COALESCE(es_moderador, FALSE) AS es_moderador,
			motivo_suspension
	`
	err := r.pool.QueryRow(ctx, query, moderator, id).Scan(&user.IDUsuario, &user.Username, &user.Correo, &user.Rol, &user.Suspendido, &user.EsModerador, &user.Motivo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
