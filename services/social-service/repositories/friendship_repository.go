package repositories

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/social-service/models"
)

var (
	ErrRequestConflict    = errors.New("friendship request conflicts with an existing relationship")
	ErrBlocked            = errors.New("friendship is blocked")
	ErrRequestNotFound    = errors.New("friendship request not found")
	ErrNotReceiver        = errors.New("user is not the request receiver")
	ErrNotSender          = errors.New("user is not the request sender")
	ErrRequestProcessed   = errors.New("friendship request was already processed")
	ErrFriendshipNotFound = errors.New("friendship not found")
	ErrSelfBlock          = errors.New("cannot block yourself")
	ErrAlreadyBlocked     = errors.New("user is already blocked")
	ErrNotFriends         = errors.New("users are not friends")
	ErrNotAuthorized      = errors.New("user is not authorized for this operation")
	ErrUserNotFound       = errors.New("user not found")
)

type FriendshipRepository interface {
	CreatePending(ctx context.Context, senderID, receiverID string) (*models.FriendRequest, error)
	ListReceived(ctx context.Context, userID string) ([]models.FriendRequest, error)
	ListSent(ctx context.Context, userID string) ([]models.FriendRequest, error)
	Accept(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error)
	Reject(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error)
	Cancel(ctx context.Context, requestID int, senderID string) (*models.FriendRequest, error)
	ListFriends(ctx context.Context, userID string) ([]models.Friend, error)
	RemoveFriend(ctx context.Context, userID, targetID string) error
	GetFriendProfile(ctx context.Context, userID, friendID string) (*models.FriendProfile, error)
	BlockUser(ctx context.Context, blockerID, blockedID string) error
	IsBlocked(ctx context.Context, user1ID, user2ID string) (bool, error)
	ValidateInteraction(ctx context.Context, user1ID, user2ID string) error
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

func (r *postgresFriendshipRepository) ListFriends(ctx context.Context, userID string) ([]models.Friend, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT 
			a.id_amistad,
			CASE WHEN a.id_usuario_1::text = $1 THEN a.id_usuario_2::text ELSE a.id_usuario_1::text END AS id_usuario,
			COALESCE(u.username, '') AS username,
			a.estado
		FROM amistad a
		LEFT JOIN usuario u ON u.id_usuario::text = (CASE WHEN a.id_usuario_1::text = $1 THEN a.id_usuario_2::text ELSE a.id_usuario_1::text END)
		WHERE (a.id_usuario_1::text = $1 OR a.id_usuario_2::text = $1)
		  AND a.estado = 'ACEPTADA'
		  AND NOT EXISTS (
			SELECT 1 FROM bloqueo b
			WHERE (b.id_usuario_bloqueador::text = $1 AND b.id_usuario_bloqueado::text = (CASE WHEN a.id_usuario_1::text = $1 THEN a.id_usuario_2::text ELSE a.id_usuario_1::text END))
			   OR (b.id_usuario_bloqueador::text = (CASE WHEN a.id_usuario_1::text = $1 THEN a.id_usuario_2::text ELSE a.id_usuario_1::text END) AND b.id_usuario_bloqueado::text = $1)
		  )
		ORDER BY a.id_amistad DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	friends := make([]models.Friend, 0)
	for rows.Next() {
		var f models.Friend
		if err := rows.Scan(&f.IDAmistad, &f.IDUsuario, &f.Username, &f.Estado); err != nil {
			return nil, err
		}
		friends = append(friends, f)
	}
	return friends, rows.Err()
}

func (r *postgresFriendshipRepository) RemoveFriend(ctx context.Context, userID, targetID string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var friendshipID int
	var u1, u2 string
	var status string

	idNum, errConv := strconv.Atoi(targetID)
	if errConv == nil && idNum > 0 {
		err = tx.QueryRow(ctx, `
			SELECT id_amistad, id_usuario_1::text, id_usuario_2::text, estado
			FROM amistad
			WHERE id_amistad = $1 FOR UPDATE`, idNum).Scan(&friendshipID, &u1, &u2, &status)
	} else {
		err = tx.QueryRow(ctx, `
			SELECT id_amistad, id_usuario_1::text, id_usuario_2::text, estado
			FROM amistad
			WHERE ((id_usuario_1::text = $1 AND id_usuario_2::text = $2)
			    OR (id_usuario_1::text = $2 AND id_usuario_2::text = $1))
			  AND estado = 'ACEPTADA'
			FOR UPDATE`, userID, targetID).Scan(&friendshipID, &u1, &u2, &status)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrFriendshipNotFound
	}
	if err != nil {
		return err
	}

	if status != models.FriendshipAccepted {
		return ErrFriendshipNotFound
	}

	if u1 != userID && u2 != userID {
		return ErrNotAuthorized
	}

	_, err = tx.Exec(ctx, `
		UPDATE amistad SET estado = $1 WHERE id_amistad = $2`,
		models.FriendshipDeleted, friendshipID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *postgresFriendshipRepository) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}
	if blockerID == blockedID {
		return ErrSelfBlock
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		SELECT pg_advisory_xact_lock(
			hashtextextended(LEAST($1::text, $2::text) || ':' || GREATEST($1::text, $2::text), 0)
		)`, blockerID, blockedID)
	if err != nil {
		return err
	}

	var alreadyBlocked bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM bloqueo
			WHERE id_usuario_bloqueador::text = $1 AND id_usuario_bloqueado::text = $2
		)`, blockerID, blockedID).Scan(&alreadyBlocked)
	if err != nil {
		return err
	}
	if alreadyBlocked {
		return ErrAlreadyBlocked
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO bloqueo (id_usuario_bloqueador, id_usuario_bloqueado)
		VALUES ($1, $2)`, blockerID, blockedID)
	if err != nil {
		return err
	}

	// Eliminar cualquier relación de amistad activa o pendiente entre ambos usuarios
	_, err = tx.Exec(ctx, `
		UPDATE amistad
		SET estado = $1
		WHERE ((id_usuario_1::text = $2 AND id_usuario_2::text = $3)
		    OR (id_usuario_1::text = $3 AND id_usuario_2::text = $2))
		  AND estado IN ('PENDIENTE', 'ACEPTADA')`,
		models.FriendshipDeleted, blockerID, blockedID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *postgresFriendshipRepository) IsBlocked(ctx context.Context, user1ID, user2ID string) (bool, error) {
	if r.pool == nil {
		return false, errors.New("database connection pool is not initialized")
	}

	var blocked bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM bloqueo
			WHERE (id_usuario_bloqueador::text = $1 AND id_usuario_bloqueado::text = $2)
			   OR (id_usuario_bloqueador::text = $2 AND id_usuario_bloqueado::text = $1)
		)`, user1ID, user2ID).Scan(&blocked)
	if err != nil {
		return false, err
	}
	return blocked, nil
}

func (r *postgresFriendshipRepository) ValidateInteraction(ctx context.Context, user1ID, user2ID string) error {
	blocked, err := r.IsBlocked(ctx, user1ID, user2ID)
	if err != nil {
		return err
	}
	if blocked {
		return ErrBlocked
	}
	return nil
}

func (r *postgresFriendshipRepository) GetFriendProfile(ctx context.Context, userID, friendID string) (*models.FriendProfile, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	// 1. Validar bloqueo bidireccional inicial
	var blocked bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM bloqueo
			WHERE (id_usuario_bloqueador::text = $1 AND id_usuario_bloqueado::text = $2)
			   OR (id_usuario_bloqueador::text = $2 AND id_usuario_bloqueado::text = $1)
		)`, userID, friendID).Scan(&blocked)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrBlocked
	}

	// 2. Verificar existencia de relación de amistad aceptada
	var friendshipID int
	var resolvedFriendID string
	idNum, errConv := strconv.Atoi(friendID)
	if errConv == nil && idNum > 0 {
		var u1, u2 string
		err = r.pool.QueryRow(ctx, `
			SELECT id_amistad, id_usuario_1::text, id_usuario_2::text
			FROM amistad
			WHERE id_amistad = $1 AND estado = 'ACEPTADA'`, idNum).Scan(&friendshipID, &u1, &u2)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFriendshipNotFound
		}
		if err != nil {
			return nil, err
		}
		if u1 != userID && u2 != userID {
			return nil, ErrNotAuthorized
		}
		if u1 == userID {
			resolvedFriendID = u2
		} else {
			resolvedFriendID = u1
		}
	} else {
		resolvedFriendID = friendID
		err = r.pool.QueryRow(ctx, `
			SELECT id_amistad
			FROM amistad
			WHERE ((id_usuario_1::text = $1 AND id_usuario_2::text = $2)
			    OR (id_usuario_1::text = $2 AND id_usuario_2::text = $1))
			  AND estado = 'ACEPTADA'`, userID, resolvedFriendID).Scan(&friendshipID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFriends
		}
		if err != nil {
			return nil, err
		}
	}

	// Comprobar bloqueo bidireccional con el ID de usuario resuelto
	if errConv == nil && idNum > 0 {
		err = r.pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM bloqueo
				WHERE (id_usuario_bloqueador::text = $1 AND id_usuario_bloqueado::text = $2)
				   OR (id_usuario_bloqueador::text = $2 AND id_usuario_bloqueado::text = $1)
			)`, userID, resolvedFriendID).Scan(&blocked)
		if err != nil {
			return nil, err
		}
		if blocked {
			return nil, ErrBlocked
		}
	}

	// 3. Obtener datos de perfil del usuario amigo
	var profile models.FriendProfile
	profile.IDAmistad = friendshipID
	err = r.pool.QueryRow(ctx, `
		SELECT id_usuario::text, username, correo, id_nivel, experiencia
		FROM usuario
		WHERE id_usuario::text = $1`, resolvedFriendID).Scan(
		&profile.IDUsuario,
		&profile.Username,
		&profile.Correo,
		&profile.IDNivel,
		&profile.Experiencia,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &profile, nil
}
