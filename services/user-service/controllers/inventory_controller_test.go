package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/user-service/controllers"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type mockInventoryService struct {
	getInventoryFunc  func(ctx context.Context, userID string) ([]*models.InventoryItemResponse, error)
	equipRewardFunc   func(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error)
	unequipRewardFunc func(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error)
}

func (m *mockInventoryService) GetInventory(ctx context.Context, userID string) ([]*models.InventoryItemResponse, error) {
	if m.getInventoryFunc != nil {
		return m.getInventoryFunc(ctx, userID)
	}
	return []*models.InventoryItemResponse{
		{
			IDInventario:    "inv-1",
			IDRecompensa:    "rew-1",
			Tipo:            "MARCO",
			TipoRecompensa:  "MARCO",
			FechaDesbloqueo: time.Now(),
			Equipada:        true,
		},
	}, nil
}

func (m *mockInventoryService) EquipReward(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error) {
	if m.equipRewardFunc != nil {
		return m.equipRewardFunc(ctx, userID, inventoryID)
	}
	return &models.InventoryItemResponse{
		IDInventario:    inventoryID,
		IDRecompensa:    "rew-1",
		Tipo:            "MARCO",
		TipoRecompensa:  "MARCO",
		FechaDesbloqueo: time.Now(),
		Equipada:        true,
	}, nil
}

func (m *mockInventoryService) UnequipReward(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error) {
	if m.unequipRewardFunc != nil {
		return m.unequipRewardFunc(ctx, userID, inventoryID)
	}
	return &models.InventoryItemResponse{
		IDInventario:    inventoryID,
		IDRecompensa:    "rew-1",
		Tipo:            "MARCO",
		TipoRecompensa:  "MARCO",
		FechaDesbloqueo: time.Now(),
		Equipada:        false,
	}, nil
}

func setupInventoryTestRouter(srv services.InventoryService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NewInventoryController(srv)

	v1Protected := router.Group("/v1/users")
	v1Protected.Use(controllers.AuthRequired())
	{
		v1Protected.GET("/me/inventory", ctrl.GetInventory)
		v1Protected.PUT("/me/inventory/:id/equip", ctrl.EquipReward)
		v1Protected.PUT("/me/inventory/:id/unequip", ctrl.UnequipReward)
	}

	protected := router.Group("/users")
	protected.Use(controllers.AuthRequired())
	{
		protected.GET("/me/inventory", ctrl.GetInventory)
		protected.PUT("/me/inventory/:id/equip", ctrl.EquipReward)
		protected.PUT("/me/inventory/:id/unequip", ctrl.UnequipReward)
	}

	return router
}

func TestGetInventory_HTTP_Exitoso(t *testing.T) {
	router := setupInventoryTestRouter(&mockInventoryService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/users/me/inventory", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error al deserializar JSON: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("se esperaba success = true")
	}

	data, ok := resp["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Fatalf("se esperaba un arreglo en 'data'")
	}

	firstItem := data[0].(map[string]interface{})
	if firstItem["id_inventario"] != "inv-1" {
		t.Errorf("id_inventario incorrecto: %v", firstItem["id_inventario"])
	}
	if firstItem["tipo"] != "MARCO" {
		t.Errorf("tipo incorrecto: %v", firstItem["tipo"])
	}
	if firstItem["equipada"] != true {
		t.Errorf("equipada incorrecto: %v", firstItem["equipada"])
	}
}

func TestGetInventory_HTTP_NoAutenticado(t *testing.T) {
	router := setupInventoryTestRouter(&mockInventoryService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/users/me/inventory", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 Unauthorized, se obtuvo %d", w.Code)
	}
}

func TestEquipReward_HTTP_Exitoso(t *testing.T) {
	router := setupInventoryTestRouter(&mockInventoryService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/inventory/inv-123/equip", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Errorf("se esperaba success = true")
	}
}

func TestEquipReward_HTTP_RecompensaDeshabilitada_RN12(t *testing.T) {
	mockSvc := &mockInventoryService{
		equipRewardFunc: func(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error) {
			return nil, services.ErrRewardDisabled
		},
	}
	router := setupInventoryTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/inventory/inv-deshab/equip", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba 400 Bad Request para recompensa deshabilitada, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != false {
		t.Errorf("se esperaba success = false")
	}
}

func TestEquipReward_HTTP_NoPertenencia_Forbidden(t *testing.T) {
	mockSvc := &mockInventoryService{
		equipRewardFunc: func(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error) {
			return nil, services.ErrItemNotOwned
		},
	}
	router := setupInventoryTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/inventory/inv-otro-usuario/equip", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 Forbidden para ítem de otro usuario, se obtuvo %d", w.Code)
	}
}

func TestEquipReward_HTTP_NoDesbloqueado_NotFound(t *testing.T) {
	mockSvc := &mockInventoryService{
		equipRewardFunc: func(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error) {
			return nil, services.ErrInventoryItemNotFound
		},
	}
	router := setupInventoryTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/inventory/inv-inexistente/equip", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404 Not Found, se obtuvo %d", w.Code)
	}
}

func TestEquipReward_HTTP_YaEquipado_Conflict(t *testing.T) {
	mockSvc := &mockInventoryService{
		equipRewardFunc: func(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error) {
			return nil, services.ErrItemAlreadyEquipped
		},
	}
	router := setupInventoryTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/inventory/inv-1/equip", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("se esperaba 409 Conflict, se obtuvo %d", w.Code)
	}
}

func TestUnequipReward_HTTP_Exitoso(t *testing.T) {
	router := setupInventoryTestRouter(&mockInventoryService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/inventory/inv-123/unequip", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Errorf("se esperaba success = true")
	}
}

func TestUnequipReward_HTTP_YaDesequipado_Conflict(t *testing.T) {
	mockSvc := &mockInventoryService{
		unequipRewardFunc: func(ctx context.Context, userID string, inventoryID string) (*models.InventoryItemResponse, error) {
			return nil, services.ErrItemAlreadyUnequipped
		},
	}
	router := setupInventoryTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/inventory/inv-1/unequip", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("se esperaba 409 Conflict, se obtuvo %d", w.Code)
	}
}
