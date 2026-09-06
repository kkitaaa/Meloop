package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/auth-service/models"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.Usuario, error)
	GetByEmail(ctx context.Context, email string) (*models.Usuario, error)
	Create(ctx context.Context, user *models.Usuario) error
}

type postgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{pool: pool}
}

func (r *postgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.Usuario, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
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

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.Usuario, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
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

func (r *postgresUserRepository) Create(ctx context.Context, user *models.Usuario) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}
	query := `
		INSERT INTO USUARIO (username, correo, contrasena_hash, id_nivel, experiencia)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id_usuario::text
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		user.Username,
		user.Correo,
		user.ContrasenaHash,
		user.IDNivel,
		user.Experiencia,
	).Scan(&user.IDUsuario)

	return err
}
