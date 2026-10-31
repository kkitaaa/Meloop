package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/services"
)

// InventoryController maneja las peticiones HTTP relacionadas con el inventario y equipamiento (RF-08, RF-09, RN-12)
type InventoryController struct {
	inventoryService services.InventoryService
}

// NewInventoryController crea una nueva instancia de InventoryController
func NewInventoryController(inventoryService services.InventoryService) *InventoryController {
	return &InventoryController{inventoryService: inventoryService}
}

// GetInventory maneja la consulta del inventario del usuario autenticado (GET /v1/users/me/inventory)
func (ctrl *InventoryController) GetInventory(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	items, err := ctrl.inventoryService.GetInventory(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}
		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, items)
}

// EquipReward maneja el equipamiento de un elemento de inventario (PUT /v1/users/me/inventory/{id}/equip)
func (ctrl *InventoryController) EquipReward(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	inventoryID := strings.TrimSpace(c.Param("id"))
	if inventoryID == "" {
		httpresponse.BadRequestGin(c, "Identificador de inventario no proporcionado")
		return
	}

	item, err := ctrl.inventoryService.EquipReward(c.Request.Context(), userID, inventoryID)
	if err != nil {
		var valErr *services.ValidationError
		if errors.As(err, &valErr) {
			details := map[string]string{
				"field": valErr.Field,
				"issue": valErr.Issue,
			}
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpresponse.ErrValidation,
				valErr.Message,
				details,
			)
			return
		}

		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}

		if errors.Is(err, services.ErrInventoryItemNotFound) || errors.Is(err, services.ErrItemNotUnlocked) {
			httpresponse.NotFoundGin(c, "Elemento de inventario no encontrado o no desbloqueado")
			return
		}

		if errors.Is(err, services.ErrItemNotOwned) {
			httpresponse.ForbiddenGin(c, "No tienes permiso para equipar este elemento de inventario")
			return
		}

		if errors.Is(err, services.ErrRewardDisabled) {
			httpresponse.BadRequestGin(c, "La recompensa se encuentra deshabilitada y no puede ser equipada")
			return
		}

		if errors.Is(err, services.ErrItemAlreadyEquipped) {
			httpresponse.ConflictGin(c, "El elemento ya se encuentra equipado")
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, item)
}

// UnequipReward maneja la desequipación de un elemento de inventario (PUT /v1/users/me/inventory/{id}/unequip)
func (ctrl *InventoryController) UnequipReward(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	inventoryID := strings.TrimSpace(c.Param("id"))
	if inventoryID == "" {
		httpresponse.BadRequestGin(c, "Identificador de inventario no proporcionado")
		return
	}

	item, err := ctrl.inventoryService.UnequipReward(c.Request.Context(), userID, inventoryID)
	if err != nil {
		var valErr *services.ValidationError
		if errors.As(err, &valErr) {
			details := map[string]string{
				"field": valErr.Field,
				"issue": valErr.Issue,
			}
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpresponse.ErrValidation,
				valErr.Message,
				details,
			)
			return
		}

		if errors.Is(err, services.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "Usuario no encontrado")
			return
		}

		if errors.Is(err, services.ErrInventoryItemNotFound) || errors.Is(err, services.ErrItemNotUnlocked) {
			httpresponse.NotFoundGin(c, "Elemento de inventario no encontrado o no desbloqueado")
			return
		}

		if errors.Is(err, services.ErrItemNotOwned) {
			httpresponse.ForbiddenGin(c, "No tienes permiso para desequipar este elemento de inventario")
			return
		}

		if errors.Is(err, services.ErrItemAlreadyUnequipped) {
			httpresponse.ConflictGin(c, "El elemento ya se encuentra desequipado")
			return
		}

		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, item)
}
