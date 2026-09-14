package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/social-service/models"
)

var (
	ErrRequestConflict  = errors.New("friendship request conflicts with an existing relationship")
	ErrBlocked          = errors.New("friendship is blocked")
	ErrRequestNotFound  = errors.New("friendship request not found")
	ErrNotReceiver      = errors.New("user is not the request receiver")
	ErrNotSender        = errors.New("user is not the request sender")
	ErrRequestProcessed = errors.New("friendship request was already processed")
)

type FriendshipRepository interface {
	CreatePending(ctx context.Context, senderID, receiverID string) (*models.FriendRequest, error)
	ListReceived(ctx context.Context, userID string) ([]models.FriendRequest, error)
	ListSent(ctx context.Context, userID string) ([]models.FriendRequest, error)
	Accept(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error)
	Reject(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error)
	Cancel(ctx context.Context, requestID int, senderID string) (*models.FriendRequest, error)
}

type postgresFriendshipRepository struct {
	pool *pgxpool.Pool
}

func NewFriendshipRepository(pool *pgxpool.Pool) FriendshipRepository {
	return &postgresFriendshipRepository{pool: pool}
}

func (r *postgresFriendshipRepository) CreatePending(ctx context.Context, senderID, receiverID string) (*models.FriendRequest, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		SELECT pg_advisory_xact_lock(
			hashtextextended(LEAST($1::text, $2::text) || ':' || GREATEST($1::text, $2::text), 0)
		)`, senderID, receiverID)
	if err != nil {
		return nil, err
	}

	var blocked bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM bloqueo
			WHERE (id_usuario_bloqueador = $1 AND id_usuario_bloqueado = $2)
			   OR (id_usuario_bloqueador = $2 AND id_usuario_bloqueado = $1)
		)`, senderID, receiverID).Scan(&blocked)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrBlocked
	}

	var reception string
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(recepcion_solicitudes_amistad, 'TODOS')
		FROM configuracion_privacidad WHERE id_usuario = $1`, receiverID).Scan(&reception)
	if errors.Is(err, pgx.ErrNoRows) {
		reception = "TODOS"
	} else if err != nil {
		return nil, err
	}
	if reception == "NADIE" || reception == "AMIGOS" {
		return nil, ErrBlocked
	}

	var request models.FriendRequest
	err = tx.QueryRow(ctx, `
		SELECT id_amistad, id_usuario_1::text, id_usuario_2::text, estado
		FROM amistad
		WHERE ((id_usuario_1 = $1 AND id_usuario_2 = $2)
		    OR (id_usuario_1 = $2 AND id_usuario_2 = $1))
		  AND estado IN ('PENDIENTE', 'ACEPTADA')
		LIMIT 1`, senderID, receiverID).Scan(
		&request.ID, &request.SenderID, &request.ReceiverID, &request.Status,
	)
	if err == nil {
		return nil, ErrRequestConflict
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO amistad (id_usuario_1, id_usuario_2, estado)
		VALUES ($1, $2, 'PENDIENTE')
		RETURNING id_amistad, id_usuario_1::text, id_usuario_2::text, estado`, senderID, receiverID).Scan(
		&request.ID, &request.SenderID, &request.ReceiverID, &request.Status,
	)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *postgresFriendshipRepository) ListReceived(ctx context.Context, userID string) ([]models.FriendRequest, error) {
	return r.list(ctx, `id_usuario_2 = $1`, userID)
}

func (r *postgresFriendshipRepository) ListSent(ctx context.Context, userID string) ([]models.FriendRequest, error) {
	return r.list(ctx, `id_usuario_1 = $1`, userID)
}

func (r *postgresFriendshipRepository) list(ctx context.Context, condition, userID string) ([]models.FriendRequest, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id_amistad, id_usuario_1::text, id_usuario_2::text, estado
		FROM amistad WHERE `+condition+` AND estado = 'PENDIENTE'
		ORDER BY id_amistad DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := make([]models.FriendRequest, 0)
	for rows.Next() {
		var request models.FriendRequest
		if err := rows.Scan(&request.ID, &request.SenderID, &request.ReceiverID, &request.Status); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

func (r *postgresFriendshipRepository) Accept(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error) {
	return r.process(ctx, requestID, receiverID, models.FriendshipAccepted, true)
}

func (r *postgresFriendshipRepository) Reject(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error) {
	return r.process(ctx, requestID, receiverID, models.FriendshipRejected, true)
}

func (r *postgresFriendshipRepository) Cancel(ctx context.Context, requestID int, senderID string) (*models.FriendRequest, error) {
	return r.process(ctx, requestID, senderID, models.FriendshipCanceled, false)
}

func (r *postgresFriendshipRepository) process(ctx context.Context, requestID int, userID, newStatus string, receiver bool) (*models.FriendRequest, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var request models.FriendRequest
	err = tx.QueryRow(ctx, `
		SELECT id_amistad, id_usuario_1::text, id_usuario_2::text, estado
		FROM amistad WHERE id_amistad = $1 FOR UPDATE`, requestID).Scan(
		&request.ID, &request.SenderID, &request.ReceiverID, &request.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRequestNotFound
	}
	if err != nil {
		return nil, err
	}
	if request.Status != models.FriendshipPending {
		return nil, ErrRequestProcessed
	}
	if (receiver && request.ReceiverID != userID) || (!receiver && request.SenderID != userID) {
		if receiver {
			return nil, ErrNotReceiver
		}
		return nil, ErrNotSender
	}

	err = tx.QueryRow(ctx, `
		UPDATE amistad SET estado = $1 WHERE id_amistad = $2
		RETURNING id_amistad, id_usuario_1::text, id_usuario_2::text, estado`, newStatus, requestID).Scan(
		&request.ID, &request.SenderID, &request.ReceiverID, &request.Status,
	)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &request, nil
}
