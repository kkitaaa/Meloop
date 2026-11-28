package repositories

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/auth-service/models"
)

// ErrUserNotFound indica que la identidad de la sesión ya no existe en PostgreSQL.
var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.Usuario, error)
	GetByUsername(ctx context.Context, username string) (*models.Usuario, error)
	GetByEmail(ctx context.Context, email string) (*models.Usuario, error)
	GetRoleByID(ctx context.Context, id string) (string, error)
	Create(ctx context.Context, user *models.Usuario) error
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id string) error
}

type postgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{pool: pool}
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id string) (*models.Usuario, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
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

// GetRoleByID devuelve el rol vigente desde USUARIO. Un usuario inexistente
// se reporta como ErrUserNotFound para que la sesión se considere inválida.
func (r *postgresUserRepository) GetRoleByID(ctx context.Context, id string) (string, error) {
	if r.pool == nil {
		return "", errors.New("database connection pool is not initialized")
	}
	var role string
	query := `
		SELECT COALESCE(rol, 'usuario')
		FROM USUARIO
		WHERE id_usuario = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return role, nil
}

func (r *postgresUserRepository) Create(ctx context.Context, user *models.Usuario) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	if user.IDUsuario == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err == nil {
			user.IDUsuario = hex.EncodeToString(b)
		} else {
			user.IDUsuario = fmt.Sprintf("u_%d", time.Now().UnixNano())
		}
	}

	query := `
		INSERT INTO USUARIO (id_usuario, username, correo, contrasena_hash, id_nivel, experiencia, fecha_creacion)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id_usuario::text
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		user.IDUsuario,
		user.Username,
		user.Correo,
		user.ContrasenaHash,
		user.IDNivel,
		user.Experiencia,
	).Scan(&user.IDUsuario)

	return err
}

func (r *postgresUserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}
	query := `
		UPDATE USUARIO
		SET contrasena_hash = $1
		WHERE id_usuario = $2
	`
	_, err := r.pool.Exec(ctx, query, passwordHash, id)
	return err
}

func (r *postgresUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}
	result, err := r.pool.Exec(ctx, `
		UPDATE USUARIO
		SET fecha_ultimo_login = NOW()
		WHERE id_usuario = $1`,
		id,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}
