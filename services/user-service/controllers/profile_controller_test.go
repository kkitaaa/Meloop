package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type mockProfileService struct {
	getProfileFunc    func(ctx context.Context, userID string) (*models.ProfileResponse, error)
	updateProfileFunc func(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.ProfileResponse, error)
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
		protected.PUT("/me/profile", ctrl.UpdateProfile)
	}

	v1Protected := r.Group("/v1/users")
	v1Protected.Use(AuthRequired())
	{
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
