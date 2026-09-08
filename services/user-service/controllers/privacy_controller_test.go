package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/user-service/controllers"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type mockPrivacyService struct {
	getPrivacyFunc    func(ctx context.Context, userID string) (*models.PrivacyResponse, error)
	updatePrivacyFunc func(ctx context.Context, userID string, req *models.UpdatePrivacyRequest) (*models.PrivacyResponse, error)
}

func (m *mockPrivacyService) GetPrivacy(ctx context.Context, userID string) (*models.PrivacyResponse, error) {
	if m.getPrivacyFunc != nil {
		return m.getPrivacyFunc(ctx, userID)
	}
	return &models.PrivacyResponse{
		IDUsuario:                   userID,
		VisibilidadPerfil:           models.VisibilidadPublico,
		VisibilidadPublicaciones:    models.VisibilidadPublico,
		VisibilidadInteracciones:    models.VisibilidadPublico,
		RecepcionMensajes:           models.RecepcionTodos,
		RecepcionSolicitudesAmistad: models.RecepcionTodos,
	}, nil
}

func (m *mockPrivacyService) UpdatePrivacy(ctx context.Context, userID string, req *models.UpdatePrivacyRequest) (*models.PrivacyResponse, error) {
	if m.updatePrivacyFunc != nil {
		return m.updatePrivacyFunc(ctx, userID, req)
	}
	vis := models.VisibilidadPublico
	if req.VisibilidadPerfil != nil {
		vis = *req.VisibilidadPerfil
	}
	return &models.PrivacyResponse{
		IDUsuario:                   userID,
		VisibilidadPerfil:           vis,
		VisibilidadPublicaciones:    models.VisibilidadPublico,
		VisibilidadInteracciones:    models.VisibilidadPublico,
		RecepcionMensajes:           models.RecepcionTodos,
		RecepcionSolicitudesAmistad: models.RecepcionTodos,
	}, nil
}

func setupPrivacyTestRouter(srv services.PrivacyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NewPrivacyController(srv)

	protected := router.Group("/users")
	protected.Use(controllers.AuthRequired())
	{
		protected.GET("/me/privacy", ctrl.GetPrivacy)
		protected.PATCH("/me/privacy", ctrl.UpdatePrivacy)
	}

	return router
}

func TestGetPrivacy_HTTP_Exitoso(t *testing.T) {
	router := setupPrivacyTestRouter(&mockPrivacyService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/privacy", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error al deserializar respuesta JSON: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("se esperaba success = true, se obtuvo %v", resp["success"])
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("se esperaba el campo 'data' en la respuesta")
	}

	if data["id_usuario"] != "usr-123" {
		t.Errorf("esperado id_usuario 'usr-123', obtenido %v", data["id_usuario"])
	}
	if data["visibilidad_perfil"] != models.VisibilidadPublico {
		t.Errorf("esperado visibilidad_perfil 'PUBLICO', obtenido %v", data["visibilidad_perfil"])
	}
}

func TestGetPrivacy_HTTP_NoAutenticado(t *testing.T) {
	router := setupPrivacyTestRouter(&mockPrivacyService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/privacy", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 Unauthorized, se obtuvo %d", w.Code)
	}
}

func TestUpdatePrivacy_HTTP_Exitoso(t *testing.T) {
	router := setupPrivacyTestRouter(&mockPrivacyService{})

	w := httptest.NewRecorder()
	body := `{"visibilidad_perfil": "PRIVADO"}`
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/privacy", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["visibilidad_perfil"] != "PRIVADO" {
		t.Errorf("se esperaba 'PRIVADO', se obtuvo %v", data["visibilidad_perfil"])
	}
}

func TestUpdatePrivacy_HTTP_ValidacionInvalida(t *testing.T) {
	mockSvc := &mockPrivacyService{
		updatePrivacyFunc: func(ctx context.Context, userID string, req *models.UpdatePrivacyRequest) (*models.PrivacyResponse, error) {
			return nil, &services.ValidationError{
				Field:   "visibilidad_perfil",
				Issue:   "invalid_value",
				Message: "Valor no permitido para visibilidad_perfil",
			}
		},
	}
	router := setupPrivacyTestRouter(mockSvc)

	w := httptest.NewRecorder()
	body := `{"visibilidad_perfil": "INVALIDO"}`
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/privacy", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba 400 Bad Request, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != false {
		t.Errorf("se esperaba success = false")
	}
}
