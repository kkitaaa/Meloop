package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/models"
)

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) IsAdmin(ctx context.Context, userID string) (bool, error) {
	if r.pool == nil {
		return false, errors.New("database connection pool is not initialized")
	}
	var isAdmin bool
	err := r.pool.QueryRow(ctx, `SELECT rol = 'ADMIN' FROM usuario WHERE id_usuario = $1`, userID).Scan(&isAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return isAdmin, err
}

func (r *AdminRepository) ListUsers(ctx context.Context) ([]models.AdminUser, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id_usuario::text, username, correo, rol, suspendido, puede_moderar
		FROM usuario
		ORDER BY username, id_usuario`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]models.AdminUser, 0)
	for rows.Next() {
		var user models.AdminUser
		if err := rows.Scan(&user.IDUsuario, &user.Username, &user.Correo, &user.Role, &user.Suspendido, &user.PuedeModerar); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *AdminRepository) SetSuspension(ctx context.Context, userID string, suspended bool) (*models.AdminUser, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	var user models.AdminUser
	err := r.pool.QueryRow(ctx, `
		UPDATE usuario
		SET suspendido = $2
		WHERE id_usuario = $1
		RETURNING id_usuario::text, username, correo, rol, suspendido, puede_moderar`, userID, suspended).Scan(
		&user.IDUsuario, &user.Username, &user.Correo, &user.Role, &user.Suspendido, &user.PuedeModerar,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *AdminRepository) SetModerator(ctx context.Context, userID string, enabled bool) (*models.AdminUser, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	var user models.AdminUser
	err := r.pool.QueryRow(ctx, `
		UPDATE usuario
		SET puede_moderar = $2
		WHERE id_usuario = $1
		RETURNING id_usuario::text, username, correo, rol, suspendido, puede_moderar`, userID, enabled).Scan(
		&user.IDUsuario, &user.Username, &user.Correo, &user.Role, &user.Suspendido, &user.PuedeModerar,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}
