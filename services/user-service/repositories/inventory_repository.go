package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/models"
)

var (
	// ErrRepoInventoryNotFound indica que el elemento de inventario no existe
	ErrRepoInventoryNotFound = errors.New("REPO_INVENTORY_NOT_FOUND")
	// ErrRepoItemNotOwned indica que el elemento pertenece a otro usuario
	ErrRepoItemNotOwned = errors.New("REPO_ITEM_NOT_OWNED")
	// ErrRepoRewardDisabled indica que la recompensa está deshabilitada (RN-12)
	ErrRepoRewardDisabled = errors.New("REPO_REWARD_DISABLED")
	// ErrRepoItemAlreadyEquipped indica que el elemento ya se encuentra equipado
	ErrRepoItemAlreadyEquipped = errors.New("REPO_ITEM_ALREADY_EQUIPPED")
	// ErrRepoItemAlreadyUnequipped indica que el elemento ya está desequipado
	ErrRepoItemAlreadyUnequipped = errors.New("REPO_ITEM_ALREADY_UNEQUIPPED")
)

// InventoryRepository define las operaciones de acceso a datos para INVENTARIO y RECOMPENSA
type InventoryRepository interface {
	GetByUserID(ctx context.Context, userID string) ([]*models.InventarioItem, error)
	GetByID(ctx context.Context, inventoryID string) (*models.InventarioItem, error)
	Equip(ctx context.Context, userID string, inventoryID string) (*models.InventarioItem, error)
	Unequip(ctx context.Context, userID string, inventoryID string) (*models.InventarioItem, error)
}

type postgresInventoryRepository struct {
	pool *pgxpool.Pool
}

// NewInventoryRepository crea una nueva instancia de InventoryRepository
func NewInventoryRepository(pool *pgxpool.Pool) InventoryRepository {
	return &postgresInventoryRepository{pool: pool}
}

// GetByUserID recupera todos los elementos de inventario desbloqueados por un usuario
func (r *postgresInventoryRepository) GetByUserID(ctx context.Context, userID string) ([]*models.InventarioItem, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	query := `
		SELECT i.id_inventario, i.id_usuario::text, i.id_recompensa,
		       COALESCE(r.tipo, ''), COALESCE(i.fecha_desbloqueo, NOW()),
		       i.equipada, COALESCE(r.habilitada, TRUE)
		FROM INVENTARIO i
		JOIN RECOMPENSA r ON i.id_recompensa = r.id_recompensa
		WHERE i.id_usuario = $1
		ORDER BY i.fecha_desbloqueo DESC, i.id_inventario ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*models.InventarioItem, 0)
	for rows.Next() {
		var item models.InventarioItem
		if err := rows.Scan(
			&item.IDInventario,
			&item.IDUsuario,
			&item.IDRecompensa,
			&item.Tipo,
			&item.FechaDesbloqueo,
			&item.Equipada,
			&item.Habilitada,
		); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// GetByID recupera un elemento específico de inventario con los datos de su recompensa
func (r *postgresInventoryRepository) GetByID(ctx context.Context, inventoryID string) (*models.InventarioItem, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	query := `
		SELECT i.id_inventario, i.id_usuario::text, i.id_recompensa,
		       COALESCE(r.tipo, ''), COALESCE(i.fecha_desbloqueo, NOW()),
		       i.equipada, COALESCE(r.habilitada, TRUE)
		FROM INVENTARIO i
		JOIN RECOMPENSA r ON i.id_recompensa = r.id_recompensa
		WHERE i.id_inventario = $1
	`
	var item models.InventarioItem
	err := r.pool.QueryRow(ctx, query, inventoryID).Scan(
		&item.IDInventario,
		&item.IDUsuario,
		&item.IDRecompensa,
		&item.Tipo,
		&item.FechaDesbloqueo,
		&item.Equipada,
		&item.Habilitada,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepoInventoryNotFound
		}
		return nil, err
	}
	return &item, nil
}

// Equip marca un elemento como equipado dentro de una transacción, desmarcando cualquier otro elemento del mismo tipo
func (r *postgresInventoryRepository) Equip(ctx context.Context, userID string, inventoryID string) (*models.InventarioItem, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Obtener y bloquear el elemento del inventario
	query := `
		SELECT i.id_inventario, i.id_usuario::text, i.id_recompensa,
		       COALESCE(r.tipo, ''), COALESCE(i.fecha_desbloqueo, NOW()),
		       i.equipada, COALESCE(r.habilitada, TRUE)
		FROM INVENTARIO i
		JOIN RECOMPENSA r ON i.id_recompensa = r.id_recompensa
		WHERE i.id_inventario = $1
		FOR UPDATE
	`
	var item models.InventarioItem
	err = tx.QueryRow(ctx, query, inventoryID).Scan(
		&item.IDInventario,
		&item.IDUsuario,
		&item.IDRecompensa,
		&item.Tipo,
		&item.FechaDesbloqueo,
		&item.Equipada,
		&item.Habilitada,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepoInventoryNotFound
		}
		return nil, err
	}

	// 2. Verificar que el elemento pertenece al usuario autenticado
	if item.IDUsuario != userID {
		return nil, ErrRepoItemNotOwned
	}

	// 3. Verificar vigencia/habilitación (RN-12)
	if !item.Habilitada {
		return nil, ErrRepoRewardDisabled
	}

	// 4. Verificar si ya está equipado
	if item.Equipada {
		return nil, ErrRepoItemAlreadyEquipped
	}

	// 5. Desequipar cualquier otro elemento del mismo tipo para este usuario
	unequipQuery := `
		UPDATE INVENTARIO
		SET equipada = FALSE
		FROM RECOMPENSA
		WHERE INVENTARIO.id_recompensa = RECOMPENSA.id_recompensa
		  AND INVENTARIO.id_usuario = $1
		  AND INVENTARIO.equipada = TRUE
		  AND RECOMPENSA.tipo = $2
		  AND INVENTARIO.id_inventario != $3
	`
	if _, err := tx.Exec(ctx, unequipQuery, userID, item.Tipo, inventoryID); err != nil {
		return nil, err
	}

	// 6. Marcar este elemento como equipado
	equipQuery := `
		UPDATE INVENTARIO
		SET equipada = TRUE
		WHERE id_inventario = $1
	`
	if _, err := tx.Exec(ctx, equipQuery, inventoryID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	item.Equipada = true
	return &item, nil
}

// Unequip marca un elemento como no equipado dentro de una transacción
func (r *postgresInventoryRepository) Unequip(ctx context.Context, userID string, inventoryID string) (*models.InventarioItem, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Obtener y bloquear el elemento del inventario
	query := `
		SELECT i.id_inventario, i.id_usuario::text, i.id_recompensa,
		       COALESCE(r.tipo, ''), COALESCE(i.fecha_desbloqueo, NOW()),
		       i.equipada, COALESCE(r.habilitada, TRUE)
		FROM INVENTARIO i
		JOIN RECOMPENSA r ON i.id_recompensa = r.id_recompensa
		WHERE i.id_inventario = $1
		FOR UPDATE
	`
	var item models.InventarioItem
	err = tx.QueryRow(ctx, query, inventoryID).Scan(
		&item.IDInventario,
		&item.IDUsuario,
		&item.IDRecompensa,
		&item.Tipo,
		&item.FechaDesbloqueo,
		&item.Equipada,
		&item.Habilitada,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepoInventoryNotFound
		}
		return nil, err
	}

	// 2. Verificar que pertenece al usuario autenticado
	if item.IDUsuario != userID {
		return nil, ErrRepoItemNotOwned
	}

	// 3. Verificar si ya está desequipado
	if !item.Equipada {
		return nil, ErrRepoItemAlreadyUnequipped
	}

	// 4. Desequipar
	unequipQuery := `
		UPDATE INVENTARIO
		SET equipada = FALSE
		WHERE id_inventario = $1
	`
	if _, err := tx.Exec(ctx, unequipQuery, inventoryID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	item.Equipada = false
	return &item, nil
}
