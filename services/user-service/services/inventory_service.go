package services

import (
	"context"
	"errors"
	"strings"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/repositories"
)

var (
	// ErrInventoryItemNotFound indica que el elemento de inventario no existe o no está desbloqueado
	ErrInventoryItemNotFound = errors.New("INVENTORY_ITEM_NOT_FOUND")
	// ErrItemNotUnlocked indica que el elemento no ha sido desbloqueado por el usuario
	ErrItemNotUnlocked = errors.New("ITEM_NOT_UNLOCKED")
	// ErrItemNotOwned indica que el elemento no pertenece al usuario autenticado
	ErrItemNotOwned = errors.New("ITEM_NOT_OWNED")
	// ErrRewardDisabled indica que la recompensa está deshabilitada y no puede equiparse (RN-12)
	ErrRewardDisabled = errors.New("REWARD_DISABLED")
	// ErrItemAlreadyEquipped indica que el elemento ya se encuentra equipado
	ErrItemAlreadyEquipped = errors.New("ITEM_ALREADY_EQUIPPED")
	// ErrItemAlreadyUnequipped indica que el elemento ya se encuentra desequipado
	ErrItemAlreadyUnequipped = errors.New("ITEM_ALREADY_UNEQUIPPED")
)

// InventoryService define la interfaz del servicio de inventario y equipamiento
type InventoryService interface {
	GetInventory(ctx context.Context, userID string) ([]*models.InventoryItemResponse, error)
	EquipReward(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error)
	UnequipReward(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error)
}

type inventoryService struct {
	inventoryRepo repositories.InventoryRepository
	userRepo      repositories.UserRepository
}

// NewInventoryService crea una nueva instancia de InventoryService
func NewInventoryService(inventoryRepo repositories.InventoryRepository, userRepo repositories.UserRepository) InventoryService {
	return &inventoryService{
		inventoryRepo: inventoryRepo,
		userRepo:      userRepo,
	}
}

// GetInventory obtiene la lista de recompensas desbloqueadas del usuario autenticado (RF-09)
func (s *inventoryService) GetInventory(ctx context.Context, userID string) ([]*models.InventoryItemResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	// Verificar existencia del usuario
	if s.userRepo != nil {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, ErrUserNotFound
		}
	}

	items, err := s.inventoryRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]*models.InventoryItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, &models.InventoryItemResponse{
			IDInventario:    item.IDInventario,
			IDRecompensa:    item.IDRecompensa,
			Tipo:            item.Tipo,
			TipoRecompensa:  item.Tipo,
			FechaDesbloqueo: item.FechaDesbloqueo,
			Equipada:        item.Equipada,
		})
	}

	return response, nil
}

// EquipReward equipa una recompensa desbloqueada y vigente, desmarcando elementos previos del mismo tipo (RF-08, RN-12)
func (s *inventoryService) EquipReward(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	inventoryID = strings.TrimSpace(inventoryID)
	if inventoryID == "" {
		return nil, &ValidationError{
			Field:   "id",
			Issue:   "required",
			Message: "El identificador del elemento de inventario es obligatorio",
		}
	}

	// Verificar existencia del usuario
	if s.userRepo != nil {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, ErrUserNotFound
		}
	}

	updatedItem, err := s.inventoryRepo.Equip(ctx, userID, inventoryID)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrRepoInventoryNotFound):
			return nil, ErrInventoryItemNotFound
		case errors.Is(err, repositories.ErrRepoItemNotOwned):
			return nil, ErrItemNotOwned
		case errors.Is(err, repositories.ErrRepoRewardDisabled):
			return nil, ErrRewardDisabled
		case errors.Is(err, repositories.ErrRepoItemAlreadyEquipped):
			return nil, ErrItemAlreadyEquipped
		default:
			return nil, err
		}
	}

	return &models.InventoryItemResponse{
		IDInventario:    updatedItem.IDInventario,
		IDRecompensa:    updatedItem.IDRecompensa,
		Tipo:            updatedItem.Tipo,
		TipoRecompensa:  updatedItem.Tipo,
		FechaDesbloqueo: updatedItem.FechaDesbloqueo,
		Equipada:        updatedItem.Equipada,
	}, nil
}

// UnequipReward desequipa una recompensa del inventario del usuario (RF-08, RN-12)
func (s *inventoryService) UnequipReward(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	inventoryID = strings.TrimSpace(inventoryID)
	if inventoryID == "" {
		return nil, &ValidationError{
			Field:   "id",
			Issue:   "required",
			Message: "El identificador del elemento de inventario es obligatorio",
		}
	}

	// Verificar existencia del usuario
	if s.userRepo != nil {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, ErrUserNotFound
		}
	}

	updatedItem, err := s.inventoryRepo.Unequip(ctx, userID, inventoryID)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrRepoInventoryNotFound):
			return nil, ErrInventoryItemNotFound
		case errors.Is(err, repositories.ErrRepoItemNotOwned):
			return nil, ErrItemNotOwned
		case errors.Is(err, repositories.ErrRepoItemAlreadyUnequipped):
			return nil, ErrItemAlreadyUnequipped
		default:
			return nil, err
		}
	}

	return &models.InventoryItemResponse{
		IDInventario:    updatedItem.IDInventario,
		IDRecompensa:    updatedItem.IDRecompensa,
		Tipo:            updatedItem.Tipo,
		TipoRecompensa:  updatedItem.Tipo,
		FechaDesbloqueo: updatedItem.FechaDesbloqueo,
		Equipada:        updatedItem.Equipada,
	}, nil
}
