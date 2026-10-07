package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/repositories"
	"github.com/meloop/user-service/services"
)

type mockInventoryRepository struct {
	items         map[string]*models.InventarioItem // key: id_inventario
	getByUserFunc func(ctx context.Context, userID string) ([]*models.InventarioItem, error)
	getByIDFunc   func(ctx context.Context, inventoryID string) (*models.InventarioItem, error)
	equipFunc     func(ctx context.Context, userID string, inventoryID string) (*models.InventarioItem, error)
	unequipFunc   func(ctx context.Context, userID string, inventoryID string) (*models.InventarioItem, error)
}

func (m *mockInventoryRepository) GetByUserID(ctx context.Context, userID string) ([]*models.InventarioItem, error) {
	if m.getByUserFunc != nil {
		return m.getByUserFunc(ctx, userID)
	}
	result := make([]*models.InventarioItem, 0)
	for _, item := range m.items {
		if item.IDUsuario == userID {
			// Clona para evitar mutaciones externas inesperadas
			copyItem := *item
			result = append(result, &copyItem)
		}
	}
	return result, nil
}

func (m *mockInventoryRepository) GetByID(ctx context.Context, inventoryID string) (*models.InventarioItem, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, inventoryID)
	}
	if item, ok := m.items[inventoryID]; ok {
		copyItem := *item
		return &copyItem, nil
	}
	return nil, repositories.ErrRepoInventoryNotFound
}

func (m *mockInventoryRepository) Equip(ctx context.Context, userID string, inventoryID string) (*models.InventarioItem, error) {
	if m.equipFunc != nil {
		return m.equipFunc(ctx, userID, inventoryID)
	}

	item, ok := m.items[inventoryID]
	if !ok {
		return nil, repositories.ErrRepoInventoryNotFound
	}
	if item.IDUsuario != userID {
		return nil, repositories.ErrRepoItemNotOwned
	}
	if !item.Habilitada {
		return nil, repositories.ErrRepoRewardDisabled
	}
	if item.Equipada {
		return nil, repositories.ErrRepoItemAlreadyEquipped
	}

	// Desequipar otros elementos del mismo tipo para este usuario
	for _, other := range m.items {
		if other.IDUsuario == userID && other.Tipo == item.Tipo && other.IDInventario != inventoryID {
			other.Equipada = false
		}
	}

	item.Equipada = true
	copyItem := *item
	return &copyItem, nil
}

func (m *mockInventoryRepository) Unequip(ctx context.Context, userID string, inventoryID string) (*models.InventarioItem, error) {
	if m.unequipFunc != nil {
		return m.unequipFunc(ctx, userID, inventoryID)
	}

	item, ok := m.items[inventoryID]
	if !ok {
		return nil, repositories.ErrRepoInventoryNotFound
	}
	if item.IDUsuario != userID {
		return nil, repositories.ErrRepoItemNotOwned
	}
	if !item.Equipada {
		return nil, repositories.ErrRepoItemAlreadyUnequipped
	}

	item.Equipada = false
	copyItem := *item
	return &copyItem, nil
}

func TestGetInventory_Exitoso(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}

	now := time.Now()
	invRepo := &mockInventoryRepository{
		items: map[string]*models.InventarioItem{
			"inv-1": {
				IDInventario:    "inv-1",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-marco-gold",
				Tipo:            "MARCO",
				FechaDesbloqueo: now,
				Equipada:        true,
				Habilitada:      true,
			},
			"inv-2": {
				IDInventario:    "inv-2",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-banner-rock",
				Tipo:            "BANNER",
				FechaDesbloqueo: now.Add(-time.Hour),
				Equipada:        false,
				Habilitada:      true,
			},
			"inv-3": {
				IDInventario:    "inv-3",
				IDUsuario:       "usr-2", // Pertenece a otro usuario
				IDRecompensa:    "rew-avatar-vip",
				Tipo:            "AVATAR",
				FechaDesbloqueo: now,
				Equipada:        true,
				Habilitada:      true,
			},
		},
	}

	svc := services.NewInventoryService(invRepo, userRepo)
	items, err := svc.GetInventory(context.Background(), "usr-1")
	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("se esperaban 2 elementos para usr-1, se obtuvieron %d", len(items))
	}

	// Comprobar que no hay elementos de usr-2
	for _, it := range items {
		if it.IDInventario == "inv-3" {
			t.Errorf("violación de aislamiento: inv-3 de usr-2 fue incluido")
		}
		if it.IDInventario == "inv-1" {
			if it.Tipo != "MARCO" || !it.Equipada || it.IDRecompensa != "rew-marco-gold" {
				t.Errorf("datos incorrectos en inv-1: %+v", it)
			}
		}
	}
}

func TestGetInventory_UsuarioNoExiste(t *testing.T) {
	userRepo := &mockUserRepository{
		users: make(map[string]*models.Usuario),
	}
	invRepo := &mockInventoryRepository{
		items: make(map[string]*models.InventarioItem),
	}

	svc := services.NewInventoryService(invRepo, userRepo)
	_, err := svc.GetInventory(context.Background(), "usr-non-existent")
	if !errors.Is(err, services.ErrUserNotFound) {
		t.Fatalf("se esperaba ErrUserNotFound, se obtuvo: %v", err)
	}
}

func TestEquipReward_Exitoso(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}

	invRepo := &mockInventoryRepository{
		items: map[string]*models.InventarioItem{
			"inv-1": {
				IDInventario:    "inv-1",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-marco-silver",
				Tipo:            "MARCO",
				FechaDesbloqueo: time.Now(),
				Equipada:        false,
				Habilitada:      true,
			},
		},
	}

	svc := services.NewInventoryService(invRepo, userRepo)
	res, err := svc.EquipReward(context.Background(), "usr-1", "inv-1")
	if err != nil {
		t.Fatalf("se esperaba equipamiento exitoso, se obtuvo error: %v", err)
	}

	if !res.Equipada {
		t.Errorf("se esperaba equipada = true")
	}
	if invRepo.items["inv-1"].Equipada != true {
		t.Errorf("el repositorio debía persistir equipada = true")
	}
}

func TestEquipReward_DesmarcaMismoTipo_ConservaOtrosTipos(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}

	invRepo := &mockInventoryRepository{
		items: map[string]*models.InventarioItem{
			"inv-marco-a": {
				IDInventario:    "inv-marco-a",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-marco-a",
				Tipo:            "MARCO",
				FechaDesbloqueo: time.Now(),
				Equipada:        true,
				Habilitada:      true,
			},
			"inv-marco-b": {
				IDInventario:    "inv-marco-b",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-marco-b",
				Tipo:            "MARCO",
				FechaDesbloqueo: time.Now(),
				Equipada:        false,
				Habilitada:      true,
			},
			"inv-banner-a": {
				IDInventario:    "inv-banner-a",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-banner-a",
				Tipo:            "BANNER",
				FechaDesbloqueo: time.Now(),
				Equipada:        true,
				Habilitada:      true,
			},
		},
	}

	svc := services.NewInventoryService(invRepo, userRepo)
	res, err := svc.EquipReward(context.Background(), "usr-1", "inv-marco-b")
	if err != nil {
		t.Fatalf("error al equipar marco-b: %v", err)
	}

	if res.IDInventario != "inv-marco-b" || !res.Equipada {
		t.Fatalf("marco-b debe estar equipado")
	}

	// Marco A debe haber quedado desmarcado
	if invRepo.items["inv-marco-a"].Equipada != false {
		t.Errorf("Marco A debía desmarcarse al equipar Marco B del mismo tipo")
	}

	// Marco B debe estar equipado
	if invRepo.items["inv-marco-b"].Equipada != true {
		t.Errorf("Marco B debía estar marcado como equipado")
	}

	// Banner A de OTRO tipo debe seguir equipado
	if invRepo.items["inv-banner-a"].Equipada != true {
		t.Errorf("Banner A (tipo distinto) no debía verse afectado y permanecer equipado")
	}
}

func TestEquipReward_ImpedirNoDesbloqueada(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	invRepo := &mockInventoryRepository{
		items: make(map[string]*models.InventarioItem),
	}

	svc := services.NewInventoryService(invRepo, userRepo)
	_, err := svc.EquipReward(context.Background(), "usr-1", "inv-no-desbloqueado")
	if !errors.Is(err, services.ErrInventoryItemNotFound) {
		t.Fatalf("se esperaba ErrInventoryItemNotFound, se obtuvo: %v", err)
	}
}

func TestEquipReward_ImpedirDeshabilitada_RN12(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	invRepo := &mockInventoryRepository{
		items: map[string]*models.InventarioItem{
			"inv-deshab": {
				IDInventario:    "inv-deshab",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-antigua",
				Tipo:            "MARCO",
				FechaDesbloqueo: time.Now(),
				Equipada:        false,
				Habilitada:      false, // Deshabilitada por admin (RN-12)
			},
		},
	}

	svc := services.NewInventoryService(invRepo, userRepo)
	_, err := svc.EquipReward(context.Background(), "usr-1", "inv-deshab")
	if !errors.Is(err, services.ErrRewardDisabled) {
		t.Fatalf("se esperaba ErrRewardDisabled para recompensa deshabilitada, se obtuvo: %v", err)
	}
}

func TestRN12_RecompensaEquipadaDeshabilitadaPermaneceEquipada_Y_PermiteDesequipar(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	// Recompensa que ya estaba equipada antes de ser deshabilitada
	invRepo := &mockInventoryRepository{
		items: map[string]*models.InventarioItem{
			"inv-legado": {
				IDInventario:    "inv-legado",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-legado",
				Tipo:            "MARCO",
				FechaDesbloqueo: time.Now().Add(-24 * time.Hour),
				Equipada:        true,
				Habilitada:      false, // Ahora deshabilitada
			},
		},
	}

	svc := services.NewInventoryService(invRepo, userRepo)

	// 1. Consultar inventario: Debe permanecer equipada y no borrarse
	items, err := svc.GetInventory(context.Background(), "usr-1")
	if err != nil {
		t.Fatalf("error al consultar inventario: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("la recompensa deshabilitada no debe eliminarse del inventario")
	}
	if !items[0].Equipada {
		t.Fatalf("la recompensa previamente equipada debe permanecer equipada")
	}

	// 2. Desequipar: Debe permitirse desequipar
	unequipRes, err := svc.UnequipReward(context.Background(), "usr-1", "inv-legado")
	if err != nil {
		t.Fatalf("debe poder desequiparse una recompensa deshabilitada: %v", err)
	}
	if unequipRes.Equipada != false {
		t.Errorf("se esperaba equipada = false tras desequipar")
	}

	// 3. Re-equipar: NO debe poder equiparse de nuevo tras estar deshabilitada
	_, err = svc.EquipReward(context.Background(), "usr-1", "inv-legado")
	if !errors.Is(err, services.ErrRewardDisabled) {
		t.Fatalf("se esperaba ErrRewardDisabled al intentar re-equipar elemento deshabilitado, se obtuvo: %v", err)
	}
}

func TestUnequipReward_Exitoso(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	invRepo := &mockInventoryRepository{
		items: map[string]*models.InventarioItem{
			"inv-1": {
				IDInventario:    "inv-1",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-1",
				Tipo:            "BANNER",
				FechaDesbloqueo: time.Now(),
				Equipada:        true,
				Habilitada:      true,
			},
		},
	}

	svc := services.NewInventoryService(invRepo, userRepo)
	res, err := svc.UnequipReward(context.Background(), "usr-1", "inv-1")
	if err != nil {
		t.Fatalf("se esperaba desequipamiento exitoso, se obtuvo error: %v", err)
	}

	if res.Equipada != false {
		t.Errorf("se esperaba equipada = false")
	}
	if invRepo.items["inv-1"].Equipada != false {
		t.Errorf("el repositorio debía persistir equipada = false")
	}
}

func TestImpedirAccesoInventarioOtroUsuario(t *testing.T) {
	user1 := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	user2 := &models.Usuario{IDUsuario: "usr-2", Username: "otro"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user1, "usr-2": user2},
	}
	invRepo := &mockInventoryRepository{
		items: map[string]*models.InventarioItem{
			"inv-otro": {
				IDInventario:    "inv-otro",
				IDUsuario:       "usr-2",
				IDRecompensa:    "rew-exclusiva",
				Tipo:            "MARCO",
				FechaDesbloqueo: time.Now(),
				Equipada:        false,
				Habilitada:      true,
			},
		},
	}

	svc := services.NewInventoryService(invRepo, userRepo)

	// usr-1 intenta equipar elemento de usr-2
	_, err := svc.EquipReward(context.Background(), "usr-1", "inv-otro")
	if !errors.Is(err, services.ErrItemNotOwned) {
		t.Fatalf("se esperaba ErrItemNotOwned al intentar equipar ítem de otro usuario, se obtuvo: %v", err)
	}

	// usr-1 intenta desequipar elemento de usr-2
	invRepo.items["inv-otro"].Equipada = true
	_, err = svc.UnequipReward(context.Background(), "usr-1", "inv-otro")
	if !errors.Is(err, services.ErrItemNotOwned) {
		t.Fatalf("se esperaba ErrItemNotOwned al intentar desequipar ítem de otro usuario, se obtuvo: %v", err)
	}
}

func TestEquipReward_YaEquipado(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	invRepo := &mockInventoryRepository{
		items: map[string]*models.InventarioItem{
			"inv-1": {
				IDInventario:    "inv-1",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-1",
				Tipo:            "MARCO",
				FechaDesbloqueo: time.Now(),
				Equipada:        true,
				Habilitada:      true,
			},
		},
	}

	svc := services.NewInventoryService(invRepo, userRepo)
	_, err := svc.EquipReward(context.Background(), "usr-1", "inv-1")
	if !errors.Is(err, services.ErrItemAlreadyEquipped) {
		t.Fatalf("se esperaba ErrItemAlreadyEquipped, se obtuvo: %v", err)
	}
}

func TestUnequipReward_YaDesequipado(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	invRepo := &mockInventoryRepository{
		items: map[string]*models.InventarioItem{
			"inv-1": {
				IDInventario:    "inv-1",
				IDUsuario:       "usr-1",
				IDRecompensa:    "rew-1",
				Tipo:            "MARCO",
				FechaDesbloqueo: time.Now(),
				Equipada:        false,
				Habilitada:      true,
			},
		},
	}

	svc := services.NewInventoryService(invRepo, userRepo)
	_, err := svc.UnequipReward(context.Background(), "usr-1", "inv-1")
	if !errors.Is(err, services.ErrItemAlreadyUnequipped) {
		t.Fatalf("se esperaba ErrItemAlreadyUnequipped, se obtuvo: %v", err)
	}
}

func TestInventory_ValidacionIDRequerido(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	invRepo := &mockInventoryRepository{}
	svc := services.NewInventoryService(invRepo, userRepo)

	_, err := svc.EquipReward(context.Background(), "usr-1", "   ")
	if err == nil {
		t.Fatal("se esperaba error de validación para ID en blanco al equipar")
	}

	_, err = svc.UnequipReward(context.Background(), "usr-1", "")
	if err == nil {
		t.Fatal("se esperaba error de validación para ID vacío al desequipar")
	}
}
