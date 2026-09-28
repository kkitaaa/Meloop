package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type mockPrivacyRepository struct {
	configs       map[string]*models.ConfiguracionPrivacidad
	getByUserFunc func(ctx context.Context, userID string) (*models.ConfiguracionPrivacidad, error)
	updateFunc    func(ctx context.Context, config *models.ConfiguracionPrivacidad) (*models.ConfiguracionPrivacidad, error)
}

func (m *mockPrivacyRepository) GetByUserID(ctx context.Context, userID string) (*models.ConfiguracionPrivacidad, error) {
	if m.getByUserFunc != nil {
		return m.getByUserFunc(ctx, userID)
	}
	if cfg, ok := m.configs[userID]; ok {
		return cfg, nil
	}
	// Si no existe, simular la creación de valores por defecto
	return m.CreateDefault(ctx, userID)
}

func (m *mockPrivacyRepository) CreateDefault(ctx context.Context, userID string) (*models.ConfiguracionPrivacidad, error) {
	cfg := &models.ConfiguracionPrivacidad{
		IDPrivacidad:                1,
		IDUsuario:                   userID,
		VisibilidadPerfil:           models.VisibilidadPublico,
		VisibilidadPublicaciones:    models.VisibilidadPublico,
		VisibilidadInteracciones:    models.VisibilidadPublico,
		RecepcionMensajes:           models.RecepcionTodos,
		RecepcionSolicitudesAmistad: models.RecepcionTodos,
	}
	if m.configs == nil {
		m.configs = make(map[string]*models.ConfiguracionPrivacidad)
	}
	m.configs[userID] = cfg
	return cfg, nil
}

func (m *mockPrivacyRepository) Update(ctx context.Context, config *models.ConfiguracionPrivacidad) (*models.ConfiguracionPrivacidad, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, config)
	}
	if m.configs == nil {
		m.configs = make(map[string]*models.ConfiguracionPrivacidad)
	}
	m.configs[config.IDUsuario] = config
	return config, nil
}

func TestGetPrivacy_Exitoso(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	privacyRepo := &mockPrivacyRepository{
		configs: map[string]*models.ConfiguracionPrivacidad{
			"usr-1": {
				IDUsuario:                   "usr-1",
				VisibilidadPerfil:           models.VisibilidadPublico,
				VisibilidadPublicaciones:    models.VisibilidadAmigos,
				VisibilidadInteracciones:    models.VisibilidadPrivado,
				RecepcionMensajes:           models.RecepcionAmigos,
				RecepcionSolicitudesAmistad: models.RecepcionTodos,
			},
		},
	}

	svc := services.NewPrivacyService(privacyRepo, userRepo)
	res, err := svc.GetPrivacy(context.Background(), "usr-1")
	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}

	if res.IDUsuario != "usr-1" {
		t.Errorf("se esperaba IDUsuario 'usr-1', se obtuvo '%s'", res.IDUsuario)
	}
	if res.VisibilidadPublicaciones != models.VisibilidadAmigos {
		t.Errorf("se esperaba visibilidad_publicaciones 'AMIGOS', se obtuvo '%s'", res.VisibilidadPublicaciones)
	}
}

func TestGetPrivacy_UsuarioNoExiste(t *testing.T) {
	userRepo := &mockUserRepository{
		users: make(map[string]*models.Usuario),
	}
	privacyRepo := &mockPrivacyRepository{}

	svc := services.NewPrivacyService(privacyRepo, userRepo)
	_, err := svc.GetPrivacy(context.Background(), "usr-non-existent")
	if !errors.Is(err, services.ErrUserNotFound) {
		t.Errorf("se esperaba ErrUserNotFound, se obtuvo: %v", err)
	}
}

func TestUpdatePrivacy_Exitoso_ParcialConservaNoEnviados(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	privacyRepo := &mockPrivacyRepository{
		configs: map[string]*models.ConfiguracionPrivacidad{
			"usr-1": {
				IDUsuario:                   "usr-1",
				VisibilidadPerfil:           models.VisibilidadPublico,
				VisibilidadPublicaciones:    models.VisibilidadPublico,
				VisibilidadInteracciones:    models.VisibilidadPublico,
				RecepcionMensajes:           models.RecepcionTodos,
				RecepcionSolicitudesAmistad: models.RecepcionTodos,
			},
		},
	}

	svc := services.NewPrivacyService(privacyRepo, userRepo)
	newVis := models.VisibilidadPrivado
	req := &models.UpdatePrivacyRequest{
		VisibilidadPerfil: &newVis,
	}

	res, err := svc.UpdatePrivacy(context.Background(), "usr-1", req)
	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}

	if res.VisibilidadPerfil != models.VisibilidadPrivado {
		t.Errorf("se esperaba visibilidad_perfil 'PRIVADO', se obtuvo '%s'", res.VisibilidadPerfil)
	}
	// Comprobar que los campos no enviados conservaron su valor
	if res.VisibilidadPublicaciones != models.VisibilidadPublico {
		t.Errorf("se esperaba visibilidad_publicaciones 'PUBLICO', se obtuvo '%s'", res.VisibilidadPublicaciones)
	}
	if res.RecepcionMensajes != models.RecepcionTodos {
		t.Errorf("se esperaba recepcion_mensajes 'TODOS', se obtuvo '%s'", res.RecepcionMensajes)
	}
}

func TestUpdatePrivacy_Exitoso_MultiplesCampos(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	privacyRepo := &mockPrivacyRepository{}

	svc := services.NewPrivacyService(privacyRepo, userRepo)
	visPub := models.VisibilidadAmigos
	visInt := models.VisibilidadPrivado
	recMsg := models.RecepcionAmigos
	recSol := models.RecepcionNadie

	req := &models.UpdatePrivacyRequest{
		VisibilidadPublicaciones:    &visPub,
		VisibilidadInteracciones:    &visInt,
		RecepcionMensajes:           &recMsg,
		RecepcionSolicitudesAmistad: &recSol,
	}

	res, err := svc.UpdatePrivacy(context.Background(), "usr-1", req)
	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}

	if res.VisibilidadPublicaciones != models.VisibilidadAmigos {
		t.Errorf("esperado AMIGOS, obtenido %s", res.VisibilidadPublicaciones)
	}
	if res.VisibilidadInteracciones != models.VisibilidadPrivado {
		t.Errorf("esperado PRIVADO, obtenido %s", res.VisibilidadInteracciones)
	}
	if res.RecepcionMensajes != models.RecepcionAmigos {
		t.Errorf("esperado AMIGOS, obtenido %s", res.RecepcionMensajes)
	}
	if res.RecepcionSolicitudesAmistad != models.RecepcionNadie {
		t.Errorf("esperado NADIE, obtenido %s", res.RecepcionSolicitudesAmistad)
	}
}

func TestUpdatePrivacy_ValoresInvalidos(t *testing.T) {
	user := &models.Usuario{IDUsuario: "usr-1", Username: "alan"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user},
	}
	privacyRepo := &mockPrivacyRepository{}
	svc := services.NewPrivacyService(privacyRepo, userRepo)

	tests := []struct {
		name        string
		req         *models.UpdatePrivacyRequest
		targetField string
	}{
		{
			name: "visibilidad_perfil invalida",
			req: func() *models.UpdatePrivacyRequest {
				s := "SOLO_YO"
				return &models.UpdatePrivacyRequest{VisibilidadPerfil: &s}
			}(),
			targetField: "visibilidad_perfil",
		},
		{
			name: "visibilidad_publicaciones invalida",
			req: func() *models.UpdatePrivacyRequest {
				s := "INVALID_VALUE"
				return &models.UpdatePrivacyRequest{VisibilidadPublicaciones: &s}
			}(),
			targetField: "visibilidad_publicaciones",
		},
		{
			name: "recepcion_mensajes invalida",
			req: func() *models.UpdatePrivacyRequest {
				s := "DESCONOCIDO"
				return &models.UpdatePrivacyRequest{RecepcionMensajes: &s}
			}(),
			targetField: "recepcion_mensajes",
		},
		{
			name: "recepcion_solicitudes invalida",
			req: func() *models.UpdatePrivacyRequest {
				s := "UNKNOWN"
				return &models.UpdatePrivacyRequest{RecepcionSolicitudesAmistad: &s}
			}(),
			targetField: "recepcion_solicitudes_amistad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.UpdatePrivacy(context.Background(), "usr-1", tt.req)
			if err == nil {
				t.Fatalf("se esperaba error de validación para %s, pero no hubo error", tt.targetField)
			}
			var valErr *services.ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("se esperaba *ValidationError, se obtuvo %T: %v", err, err)
			}
			if valErr.Field != tt.targetField {
				t.Errorf("se esperaba error en campo '%s', se obtuvo en '%s'", tt.targetField, valErr.Field)
			}
		})
	}
}

func TestUpdatePrivacy_AislamientoEntreUsuarios(t *testing.T) {
	user1 := &models.Usuario{IDUsuario: "usr-1", Username: "user1"}
	user2 := &models.Usuario{IDUsuario: "usr-2", Username: "user2"}
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-1": user1, "usr-2": user2},
	}
	privacyRepo := &mockPrivacyRepository{
		configs: map[string]*models.ConfiguracionPrivacidad{
			"usr-1": {
				IDUsuario:         "usr-1",
				VisibilidadPerfil: models.VisibilidadPublico,
			},
			"usr-2": {
				IDUsuario:         "usr-2",
				VisibilidadPerfil: models.VisibilidadPublico,
			},
		},
	}

	svc := services.NewPrivacyService(privacyRepo, userRepo)
	newVis := models.VisibilidadPrivado
	_, err := svc.UpdatePrivacy(context.Background(), "usr-1", &models.UpdatePrivacyRequest{
		VisibilidadPerfil: &newVis,
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	// Comprobar que usr-2 no fue modificado
	cfg2, _ := svc.GetPrivacy(context.Background(), "usr-2")
	if cfg2.VisibilidadPerfil != models.VisibilidadPublico {
		t.Errorf("violación de aislamiento: usr-2 fue alterado al modificar usr-1")
	}
}
