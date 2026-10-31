package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/auth-service/models"
)

// PasswordRecoveryRepository define las operaciones de persistencia para tokens de recuperación
type PasswordRecoveryRepository interface {
	CreateToken(ctx context.Context, token *models.PasswordRecoveryToken) error
	GetTokenByHash(ctx context.Context, tokenHash string) (*models.PasswordRecoveryToken, error)
	MarkTokenAsUsed(ctx context.Context, idRecuperacion string) error
}

type postgresPasswordRecoveryRepository struct {
	pool *pgxpool.Pool
}

// NewPasswordRecoveryRepository crea una nueva instancia de PasswordRecoveryRepository para PostgreSQL
func NewPasswordRecoveryRepository(pool *pgxpool.Pool) PasswordRecoveryRepository {
	return &postgresPasswordRecoveryRepository{pool: pool}
}

func (r *postgresPasswordRecoveryRepository) CreateToken(ctx context.Context, token *models.PasswordRecoveryToken) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	query := `
		INSERT INTO TOKEN_RECUPERACION (id_recuperacion, id_usuario, token_hash, expira_en, usado, creado_en)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		token.IDRecuperacion,
		token.IDUsuario,
		token.TokenHash,
		token.ExpiraEn,
		token.Usado,
		token.CreadoEn,
	)
	return err
}

func (r *postgresPasswordRecoveryRepository) GetTokenByHash(ctx context.Context, tokenHash string) (*models.PasswordRecoveryToken, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	var token models.PasswordRecoveryToken
	query := `
		SELECT id_recuperacion::text, id_usuario::text, token_hash, expira_en, usado, usado_en, creado_en
		FROM TOKEN_RECUPERACION
		WHERE token_hash = $1
	`
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.IDRecuperacion,
		&token.IDUsuario,
		&token.TokenHash,
		&token.ExpiraEn,
		&token.Usado,
		&token.UsadoEn,
		&token.CreadoEn,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *postgresPasswordRecoveryRepository) MarkTokenAsUsed(ctx context.Context, idRecuperacion string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	query := `
		UPDATE TOKEN_RECUPERACION
		SET usado = TRUE, usado_en = NOW()
		WHERE id_recuperacion = $1
	`
	_, err := r.pool.Exec(ctx, query, idRecuperacion)
	return err
}
