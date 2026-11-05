package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type mockProfileService struct {
	getProfileFunc           func(ctx context.Context, userID string) (*models.ProfileResponse, error)
	updateProfileFunc        func(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.ProfileResponse, error)
	getOwnProfileFunc        func(ctx context.Context, userID string) (*models.UserProfileResponse, error)
	getThirdPartyProfileFunc func(ctx context.Context, requestingUserID, targetUserID string) (*models.ThirdPartyProfileResponse, error)
}

func (m *mockProfileService) GetProfile(ctx context.Context, userID string) (*models.ProfileResponse, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, userID)
	}
	return &models.ProfileResponse{
		IDUsuario:  userID,
		Username:   "usuario_test",
		Biografia:  "Bio actual",
		FotoPerfil: "profiles/u1/foto.png",
		Banner:     "banners/u1/banner.png",
		Tema:       "oscuro",
		Colores:    map[string]interface{}{"primario": "#1ABC9C"},
	}, nil
}

func (m *mockProfileService) GetOwnProfile(ctx context.Context, userID string) (*models.UserProfileResponse, error) {
	if m.getOwnProfileFunc != nil {
		return m.getOwnProfileFunc(ctx, userID)
	}
	lvl := 3
	return &models.UserProfileResponse{
		IDUsuario:          userID,
		Username:           "usuario_test",
		Biografia:          "Mi biografia propia",
		FotoPerfil:         "profiles/u1/foto.png",
		Banner:             "banners/u1/banner.png",
		Nivel:              &lvl,
		Experiencia:        350,
		Tema:               "oscuro",
		Colores:            map[string]interface{}{"primario": "#1ABC9C"},
		Insignias:          []models.InsigniaDTO{{IDRecompensa: "rew-ins-1", Tipo: "INSIGNIA", Equipada: true}},
		ElementosEquipados: []models.EquippedItemDTO{{IDInventario: "inv-1", IDRecompensa: "rew-marco-gold", Tipo: "MARCO"}},
		InformacionMusical: &models.MusicalPreferences{Generos: []string{"rock", "metal"}},
		Estadisticas:       &models.ProfileStatsDTO{Publicaciones: 10, Amigos: 4},
	}, nil
}

func (m *mockProfileService) GetThirdPartyProfile(ctx context.Context, requestingUserID, targetUserID string) (*models.ThirdPartyProfileResponse, error) {
	if m.getThirdPartyProfileFunc != nil {
		return m.getThirdPartyProfileFunc(ctx, requestingUserID, targetUserID)
	}
	lvl := 2
	return &models.ThirdPartyProfileResponse{
		IDUsuario:          targetUserID,
		Username:           "usuario_tercero",
		Biografia:          "Bio de tercero",
		FotoPerfil:         "profiles/u2/foto.png",
		Banner:             "banners/u2/banner.png",
		Nivel:              &lvl,
		Experiencia:        200,
		Tema:               "default",
		Colores:            map[string]interface{}{},
		Insignias:          []models.InsigniaDTO{},
		ElementosEquipados: []models.EquippedItemDTO{},
		InformacionMusical: &models.MusicalPreferences{Generos: []string{"pop"}},
		Estadisticas:       &models.ProfileStatsDTO{Publicaciones: 2, Amigos: 1},
		EsAmigo:            true,
		PerfilPrivado:      false,
	}, nil
}

func (m *mockProfileService) UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.ProfileResponse, error) {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, userID, req)
	}
	bio := ""
	if req.GetBiografia() != nil {
		bio = *req.GetBiografia()
	}
	tema := "default"
	if req.GetTema() != nil {
		tema = *req.GetTema()
	}
	foto := ""
	if req.GetFotoPerfil() != nil {
		foto = *req.GetFotoPerfil()
	}
	banner := ""
	if req.GetBanner() != nil {
		banner = *req.GetBanner()
	}

	return &models.ProfileResponse{
		IDUsuario:          userID,
		Username:           "usuario_test",
		Biografia:          bio,
		FotoPerfil:         foto,
		Banner:             banner,
		Tema:               tema,
		Colores:            req.GetColores(),
		InformacionMusical: req.GetInformacionMusical(),
	}, nil
}

func setupProfileTestRouter(srv services.ProfileService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := NewProfileController(srv)

	protected := r.Group("/users")
	protected.Use(AuthRequired())
	{
		protected.GET("/me/profile", ctrl.GetOwnProfile)
		protected.GET("/me/perfil", ctrl.GetOwnProfile)
		protected.GET("/:id/profile", ctrl.GetUserProfile)
		protected.GET("/:id/perfil", ctrl.GetUserProfile)
		protected.PUT("/me/profile", ctrl.UpdateProfile)
	}

	v1Protected := r.Group("/v1/users")
	v1Protected.Use(AuthRequired())
	{
		v1Protected.GET("/me/profile", ctrl.GetOwnProfile)
		v1Protected.GET("/me/perfil", ctrl.GetOwnProfile)
		v1Protected.GET("/:id/profile", ctrl.GetUserProfile)
		v1Protected.GET("/:id/perfil", ctrl.GetUserProfile)
		v1Protected.PUT("/me/profile", ctrl.UpdateProfile)
	}

	return r
}

func TestProfileController_UpdateProfile_PUT_V1_200(t *testing.T) {
	srv := &mockProfileService{}
	router := setupProfileTestRouter(srv)

	bio := "Nueva biografía para pruebas RF-07"
	tema := "cyberpunk"
	foto := "profiles/usr-123/foto.jpg"
	banner := "banners/usr-123/banner.png"

	body, _ := json.Marshal(models.UpdateProfileRequest{
		Biografia:  &bio,
		Tema:       &tema,
		FotoPerfil: &foto,
		Banner:     &banner,
		Colores: map[string]interface{}{
			"primario": "#FF0055",
		},
		InformacionMusical: &models.MusicalPreferences{
			Generos:  []string{"synthwave", "retrowave"},
			Artistas: []string{"kavinsky"},
		},
	})

	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200, se obtuvo: %d, cuerpo: %s", w.Code, w.Body.String())
	}

	var resp httpresponse.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error al deserializar respuesta: %v", err)
	}

	if !resp.Success {
		t.Error("se esperaba success=true")
	}
}

func TestProfileController_UpdateProfile_NoAutenticado_401(t *testing.T) {
	srv := &mockProfileService{}
	router := setupProfileTestRouter(srv)

	bio := "Intento sin sesion"
	body, _ := json.Marshal(models.UpdateProfileRequest{Biografia: &bio})

	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// Sin cabecera X-User-ID

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba código 401, se obtuvo: %d", w.Code)
	}
}

func TestProfileController_UpdateProfile_Validacion_400(t *testing.T) {
	srv := &mockProfileService{
		updateProfileFunc: func(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.ProfileResponse, error) {
			return nil, &services.ValidationError{
				Field:   "biografia",
				Issue:   "too_long",
				Message: "La biografía no puede superar los 500 caracteres",
			}
		},
	}
	router := setupProfileTestRouter(srv)

	bio := "Bio demasiado larga"
	body, _ := json.Marshal(models.UpdateProfileRequest{Biografia: &bio})

	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba código 400, se obtuvo: %d", w.Code)
	}

	var resp httpresponse.Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error == nil || resp.Error.Code != httpresponse.ErrValidation {
		t.Errorf("error esperado VALIDATION_ERROR, se obtuvo: %+v", resp.Error)
	}
}

func TestProfileController_UpdateProfile_FalloAlmacenamiento_502(t *testing.T) {
	srv := &mockProfileService{
		updateProfileFunc: func(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.ProfileResponse, error) {
			return nil, services.ErrMediaStorageFailed
		},
	}
	router := setupProfileTestRouter(srv)

	foto := "profiles/usr-123/foto.png"
	body, _ := json.Marshal(models.UpdateProfileRequest{FotoPerfil: &foto})

	req, _ := http.NewRequest(http.MethodPut, "/v1/users/me/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("se esperaba código 502 ante fallo de MinIO/S3, se obtuvo: %d", w.Code)
	}
}

// =========================================================================
// Tests para Consulta de Perfiles Propios y de Terceros (RF-06 / RN-09 / RF-55 / RNF-11)
// =========================================================================

func TestProfileController_GetOwnProfile_Success_200(t *testing.T) {
	srv := &mockProfileService{}
	router := setupProfileTestRouter(srv)

	req, _ := http.NewRequest(http.MethodGet, "/v1/users/me/profile", nil)
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200, se obtuvo: %d, cuerpo: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                       `json:"success"`
		Data    models.UserProfileResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error al deserializar respuesta: %v", err)
	}

	if !resp.Success {
		t.Error("se esperaba success=true")
	}
	if resp.Data.IDUsuario != "usr-123" {
		t.Errorf("esperado IDUsuario usr-123, obtenido %s", resp.Data.IDUsuario)
	}
	if resp.Data.Username != "usuario_test" {
		t.Errorf("esperado username usuario_test, obtenido %s", resp.Data.Username)
	}
	if resp.Data.Nivel == nil || *resp.Data.Nivel != 3 {
		t.Errorf("esperado nivel 3, obtenido %v", resp.Data.Nivel)
	}
	if resp.Data.Experiencia != 350 {
		t.Errorf("esperada experiencia 350, obtenida %d", resp.Data.Experiencia)
	}
	if len(resp.Data.Insignias) != 1 {
		t.Errorf("se esperaba 1 insignia, obtenidas %d", len(resp.Data.Insignias))
	}
	if len(resp.Data.ElementosEquipados) != 1 {
		t.Errorf("se esperaba 1 elemento equipado, obtenidos %d", len(resp.Data.ElementosEquipados))
	}
	if resp.Data.InformacionMusical == nil || len(resp.Data.InformacionMusical.Generos) != 2 {
		t.Errorf("información musical no coincide")
	}
	if resp.Data.Estadisticas == nil || resp.Data.Estadisticas.Publicaciones != 10 {
		t.Errorf("estadísticas no coinciden")
	}
}

func TestProfileController_GetOwnProfile_Unauthorized_401(t *testing.T) {
	srv := &mockProfileService{}
	router := setupProfileTestRouter(srv)

	req, _ := http.NewRequest(http.MethodGet, "/v1/users/me/profile", nil)
	// Sin cabecera X-User-ID

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba código 401 para petición sin autenticación, se obtuvo: %d", w.Code)
	}
}

func TestProfileController_GetOwnProfile_NotFound_404(t *testing.T) {
	srv := &mockProfileService{
		getOwnProfileFunc: func(ctx context.Context, userID string) (*models.UserProfileResponse, error) {
			return nil, services.ErrUserNotFound
		},
	}
	router := setupProfileTestRouter(srv)

	req, _ := http.NewRequest(http.MethodGet, "/v1/users/me/profile", nil)
	req.Header.Set("X-User-ID", "usr-inexistente")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba código 404, se obtuvo: %d", w.Code)
	}
}

func TestProfileController_GetUserProfile_Success_200(t *testing.T) {
	srv := &mockProfileService{}
	router := setupProfileTestRouter(srv)

	req, _ := http.NewRequest(http.MethodGet, "/v1/users/usr-456/profile", nil)
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200, se obtuvo: %d, cuerpo: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                             `json:"success"`
		Data    models.ThirdPartyProfileResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error al deserializar respuesta: %v", err)
	}

	if !resp.Success {
		t.Error("se esperaba success=true")
	}
	if resp.Data.IDUsuario != "usr-456" {
		t.Errorf("esperado IDUsuario usr-456, obtenido %s", resp.Data.IDUsuario)
	}
	if resp.Data.Username != "usuario_tercero" {
		t.Errorf("esperado username usuario_tercero, obtenido %s", resp.Data.Username)
	}
	if !resp.Data.EsAmigo {
		t.Error("esperado EsAmigo=true")
	}
	if resp.Data.PerfilPrivado {
		t.Error("esperado PerfilPrivado=false")
	}
}

func TestProfileController_GetUserProfile_PrivacyRespected_200(t *testing.T) {
	lvl := 5
	srv := &mockProfileService{
		getThirdPartyProfileFunc: func(ctx context.Context, requestingUserID, targetUserID string) (*models.ThirdPartyProfileResponse, error) {
			return &models.ThirdPartyProfileResponse{
				IDUsuario:     targetUserID,
				Username:      "usuario_privado",
				FotoPerfil:    "profiles/u3/foto.png",
				Nivel:         &lvl,
				Experiencia:   1200,
				EsAmigo:       false,
				PerfilPrivado: true,
			}, nil
		},
	}
	router := setupProfileTestRouter(srv)

	req, _ := http.NewRequest(http.MethodGet, "/v1/users/usr-789/profile", nil)
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200, se obtuvo: %d", w.Code)
	}

	var resp struct {
		Success bool                             `json:"success"`
		Data    models.ThirdPartyProfileResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Data.PerfilPrivado {
		t.Error("se esperaba perfil_privado=true")
	}
	if resp.Data.Biografia != "" {
		t.Errorf("la biografía no debió exponerse en perfil privado, se obtuvo: %s", resp.Data.Biografia)
	}
	if resp.Data.InformacionMusical != nil {
		t.Error("la información musical no debió exponerse en perfil privado")
	}
	if resp.Data.Estadisticas != nil {
		t.Error("las estadísticas no debieron exponerse en perfil privado")
	}
}

func TestProfileController_GetUserProfile_Blocked_404(t *testing.T) {
	srv := &mockProfileService{
		getThirdPartyProfileFunc: func(ctx context.Context, requestingUserID, targetUserID string) (*models.ThirdPartyProfileResponse, error) {
			return nil, services.ErrBlocked
		},
	}
	router := setupProfileTestRouter(srv)

	req, _ := http.NewRequest(http.MethodGet, "/v1/users/usr-blocked/profile", nil)
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba código 404 ante bloqueo (RN-09), se obtuvo: %d", w.Code)
	}

	var resp httpresponse.Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error == nil || resp.Error.Code != httpresponse.ErrNotFound {
		t.Errorf("se esperaba código de error NOT_FOUND, se obtuvo: %+v", resp.Error)
	}
}

func TestProfileController_GetUserProfile_NotFound_404(t *testing.T) {
	srv := &mockProfileService{
		getThirdPartyProfileFunc: func(ctx context.Context, requestingUserID, targetUserID string) (*models.ThirdPartyProfileResponse, error) {
			return nil, services.ErrUserNotFound
		},
	}
	router := setupProfileTestRouter(srv)

	req, _ := http.NewRequest(http.MethodGet, "/v1/users/usr-inexistente/profile", nil)
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba código 404 para usuario inexistente, se obtuvo: %d", w.Code)
	}
}

func TestProfileController_GetUserProfile_SelfID_DelegatesToOwn_200(t *testing.T) {
	srv := &mockProfileService{}
	router := setupProfileTestRouter(srv)

	// Consultar con el propio ID en la ruta parametrizada
	req, _ := http.NewRequest(http.MethodGet, "/v1/users/usr-123/profile", nil)
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200, se obtuvo: %d", w.Code)
	}

	var resp struct {
		Success bool                       `json:"success"`
		Data    models.UserProfileResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.IDUsuario != "usr-123" {
		t.Errorf("se esperaba el perfil propio del usuario autenticado")
	}
}

func TestProfileController_SensitiveDataNotExposed_RNF11(t *testing.T) {
	srv := &mockProfileService{}
	router := setupProfileTestRouter(srv)

	// 1. Probar perfil propio
	reqOwn, _ := http.NewRequest(http.MethodGet, "/v1/users/me/profile", nil)
	reqOwn.Header.Set("X-User-ID", "usr-123")
	wOwn := httptest.NewRecorder()
	router.ServeHTTP(wOwn, reqOwn)

	bodyOwn := wOwn.Body.String()
	for _, sensitive := range []string{"correo", "email", "contrasena", "password", "hash", "fecha_nacimiento", "reporte"} {
		if strings.Contains(strings.ToLower(bodyOwn), sensitive) {
			t.Fatalf("RNF-11 violado: la respuesta propia contiene '%s': %s", sensitive, bodyOwn)
		}
	}

	// 2. Probar perfil de terceros
	reqThird, _ := http.NewRequest(http.MethodGet, "/v1/users/usr-456/profile", nil)
	reqThird.Header.Set("X-User-ID", "usr-123")
	wThird := httptest.NewRecorder()
	router.ServeHTTP(wThird, reqThird)

	bodyThird := wThird.Body.String()
	for _, sensitive := range []string{"correo", "email", "contrasena", "password", "hash", "fecha_nacimiento", "reporte"} {
		if strings.Contains(strings.ToLower(bodyThird), sensitive) {
			t.Fatalf("RNF-11 violado: la respuesta a terceros contiene '%s': %s", sensitive, bodyThird)
		}
	}
}
