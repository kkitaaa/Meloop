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
