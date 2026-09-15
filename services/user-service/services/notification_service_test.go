package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type mockNotificationConfigRepository struct {
	configs       map[string]map[string]bool // userID -> tipo -> habilitada
	getByUserFunc func(ctx context.Context, userID string) ([]models.ConfiguracionNotificacion, error)
	upsertFunc    func(ctx context.Context, userID string, tipo string, habilitada bool) (*models.ConfiguracionNotificacion, error)
}

func (m *mockNotificationConfigRepository) ensureInit(userID string) {
	if m.configs == nil {
		m.configs = make(map[string]map[string]bool)
	}
	if m.configs[userID] == nil {
		m.configs[userID] = make(map[string]bool)
		for _, t := range models.AllNotificationTypes {
			m.configs[userID][t] = true
		}
	}
}

func (m *mockNotificationConfigRepository) GetByUserID(ctx context.Context, userID string) ([]models.ConfiguracionNotificacion, error) {
	if m.getByUserFunc != nil {
		return m.getByUserFunc(ctx, userID)
	}
	m.ensureInit(userID)
	var list []models.ConfiguracionNotificacion
	for i, t := range models.AllNotificationTypes {
		list = append(list, models.ConfiguracionNotificacion{
			IDConfiguracion:  i + 1,
			IDUsuario:        userID,
			TipoNotificacion: t,
			Habilitada:       m.configs[userID][t],
		})
	}
	return list, nil
}

func (m *mockNotificationConfigRepository) CreateDefaults(ctx context.Context, userID string) ([]models.ConfiguracionNotificacion, error) {
	m.ensureInit(userID)
	return m.GetByUserID(ctx, userID)
}

func (m *mockNotificationConfigRepository) Upsert(ctx context.Context, userID string, tipo string, habilitada bool) (*models.ConfiguracionNotificacion, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, userID, tipo, habilitada)
	}
	m.ensureInit(userID)
	m.configs[userID][tipo] = habilitada
	return &models.ConfiguracionNotificacion{
		IDConfiguracion:  1,
		IDUsuario:        userID,
		TipoNotificacion: tipo,
		Habilitada:       habilitada,
	}, nil
}

func (m *mockNotificationConfigRepository) UpsertBatch(ctx context.Context, userID string, items []models.NotificationSettingItem) ([]models.ConfiguracionNotificacion, error) {
	m.ensureInit(userID)
	for _, item := range items {
		if item.Habilitada != nil {
			m.configs[userID][item.TipoNotificacion] = *item.Habilitada
		}
	}
	return m.GetByUserID(ctx, userID)
}

func TestGetSettings_Exitoso(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	notifRepo := &mockNotificationConfigRepository{}

	svc := services.NewNotificationConfigService(notifRepo, userRepo)
	res, err := svc.GetSettings(context.Background(), "usr-1")
	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}

	if len(res) != len(models.AllNotificationTypes) {
		t.Errorf("se esperaban %d tipos de notificación, se obtuvieron %d", len(models.AllNotificationTypes), len(res))
	}
}

func TestGetSettings_UsuarioNoExiste(t *testing.T) {
	userRepo := &mockUserRepository{
		users: make(map[string]*models.Usuario),
	}
	notifRepo := &mockNotificationConfigRepository{}

	svc := services.NewNotificationConfigService(notifRepo, userRepo)
	_, err := svc.GetSettings(context.Background(), "usr-non-existent")
	if !errors.Is(err, services.ErrUserNotFound) {
		t.Errorf("se esperaba ErrUserNotFound, se obtuvo: %v", err)
	}
}

func TestUpdateSetting_HabilitarYDeshabilitar(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	notifRepo := &mockNotificationConfigRepository{}

	svc := services.NewNotificationConfigService(notifRepo, userRepo)

	// Deshabilitar LIKE
	res, err := svc.UpdateSetting(context.Background(), "usr-1", "LIKE", false)
	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}
	if res.TipoNotificacion != "LIKE" || res.Habilitada != false {
		t.Errorf("esperado LIKE: false, obtenido: %s: %v", res.TipoNotificacion, res.Habilitada)
	}

	// Habilitar LIKE
	res, err = svc.UpdateSetting(context.Background(), "usr-1", "LIKE", true)
	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}
	if res.TipoNotificacion != "LIKE" || res.Habilitada != true {
		t.Errorf("esperado LIKE: true, obtenido: %s: %v", res.TipoNotificacion, res.Habilitada)
	}
}

func TestUpdateSetting_TipoInvalido(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	notifRepo := &mockNotificationConfigRepository{}

	svc := services.NewNotificationConfigService(notifRepo, userRepo)
	_, err := svc.UpdateSetting(context.Background(), "usr-1", "TIPO_INEXISTENTE", true)
	if err == nil {
		t.Fatal("se esperaba error para tipo inválido")
	}

	var valErr *services.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("se esperaba *ValidationError, se obtuvo: %v", err)
	}
	if valErr.Field != "tipo_notificacion" {
		t.Errorf("esperado campo 'tipo_notificacion', obtenido: %s", valErr.Field)
	}
}

func TestUpdateBatch_Exitoso(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	notifRepo := &mockNotificationConfigRepository{}

	svc := services.NewNotificationConfigService(notifRepo, userRepo)
	falseVal := false
	trueVal := true

	items := []models.NotificationSettingItem{
		{TipoNotificacion: "LIKE", Habilitada: &falseVal},
		{TipoNotificacion: "COMENTARIO", Habilitada: &trueVal},
	}

	res, err := svc.UpdateBatch(context.Background(), "usr-1", items)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	for _, item := range res {
		if item.TipoNotificacion == "LIKE" && item.Habilitada != false {
			t.Errorf("esperado LIKE=false")
		}
		if item.TipoNotificacion == "COMENTARIO" && item.Habilitada != true {
			t.Errorf("esperado COMENTARIO=true")
		}
	}
}

func TestUpdateBatch_ErroresValidacion(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	notifRepo := &mockNotificationConfigRepository{}
	svc := services.NewNotificationConfigService(notifRepo, userRepo)

	// Lote vacío
	_, err := svc.UpdateBatch(context.Background(), "usr-1", []models.NotificationSettingItem{})
	if err == nil {
		t.Fatal("se esperaba error para lote vacío")
	}

	// Tipo inválido dentro del lote
	fVal := false
	_, err = svc.UpdateBatch(context.Background(), "usr-1", []models.NotificationSettingItem{
		{TipoNotificacion: "INVALID_TYPE", Habilitada: &fVal},
	})
	if err == nil {
		t.Fatal("se esperaba error para tipo inválido en lote")
	}

	// Habilitada nula
	_, err = svc.UpdateBatch(context.Background(), "usr-1", []models.NotificationSettingItem{
		{TipoNotificacion: "LIKE", Habilitada: nil},
	})
	if err == nil {
		t.Fatal("se esperaba error para campo habilitada nulo")
	}
}

func TestNotificationSettings_AislamientoEntreUsuarios(t *testing.T) {
	user1 := &models.Usuario{IDUsuario: "usr-1", Username: "user1"}
	user2 := &models.Usuario{IDUsuario: "usr-2", Username: "user2"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user1, "usr-2": user2},
	}
	notifRepo := &mockNotificationConfigRepository{}

	svc := services.NewNotificationConfigService(notifRepo, userRepo)

	// Deshabilitar LIKE solo para usr-1
	_, err := svc.UpdateSetting(context.Background(), "usr-1", "LIKE", false)
	if err != nil {
		t.Fatalf("error al actualizar usr-1: %v", err)
	}

	// Comprobar que usr-2 conserva LIKE = true
	res2, err := svc.GetSettings(context.Background(), "usr-2")
	if err != nil {
		t.Fatalf("error al obtener usr-2: %v", err)
	}

	for _, item := range res2 {
		if item.TipoNotificacion == "LIKE" && item.Habilitada != true {
			t.Errorf("violación de aislamiento: LIKE de usr-2 fue alterado al modificar usr-1")
		}
	}
}
