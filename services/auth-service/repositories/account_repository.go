package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/auth-service/models"
)

// AccountRepository defines the database operations for accounts
type AccountRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.Account, error)
	GetByEmail(ctx context.Context, email string) (*models.Account, error)
	Create(ctx context.Context, account *models.Account) error
}

type postgresAccountRepository struct {
	pool *pgxpool.Pool
}

// NewAccountRepository creates a PostgreSQL implementation of AccountRepository
func NewAccountRepository(pool *pgxpool.Pool) AccountRepository {
	return &postgresAccountRepository{pool: pool}
}

func (r *postgresAccountRepository) GetByUsername(ctx context.Context, username string) (*models.Account, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	var acc models.Account
	query := "SELECT id, username, email, password_hash, created_at, updated_at FROM accounts WHERE username = $1"
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&acc.ID,
		&acc.Username,
		&acc.Email,
		&acc.PasswordHash,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &acc, nil
}

func (r *postgresAccountRepository) GetByEmail(ctx context.Context, email string) (*models.Account, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	var acc models.Account
	query := "SELECT id, username, email, password_hash, created_at, updated_at FROM accounts WHERE email = $1"
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&acc.ID,
		&acc.Username,
		&acc.Email,
		&acc.PasswordHash,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &acc, nil
}

func (r *postgresAccountRepository) Create(ctx context.Context, account *models.Account) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}
	query := `
		INSERT INTO accounts (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, account.Username, account.Email, account.PasswordHash).Scan(
		&account.ID,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	return err
}
