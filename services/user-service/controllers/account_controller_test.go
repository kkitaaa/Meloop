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

// mockUserService implementa la interfaz UserService para pruebas HTTP
type mockUserService struct {
	getAccountFunc         func(ctx context.Context, userID string) (*models.AccountResponse, error)
	updateUsernameFunc     func(ctx context.Context, userID string, req *models.UpdateUsernameRequest) (*models.AccountResponse, error)
	requestEmailChangeFunc func(ctx context.Context, userID string, req *models.RequestEmailChangeRequest) (*models.EmailChangeResponse, error)
}

func (m *mockUserService) GetAccount(ctx context.Context, userID string) (*models.AccountResponse, error) {
	if m.getAccountFunc != nil {
		return m.getAccountFunc(ctx, userID)
	}
	return &models.AccountResponse{
		IDUsuario: userID,
		Username:  "alanp",
		Correo:    "alan@example.com",
	}, nil
}

func (m *mockUserService) UpdateUsername(ctx context.Context, userID string, req *models.UpdateUsernameRequest) (*models.AccountResponse, error) {
	if m.updateUsernameFunc != nil {
		return m.updateUsernameFunc(ctx, userID, req)
	}
	return &models.AccountResponse{
		IDUsuario: userID,
		Username:  req.Username,
		Correo:    "alan@example.com",
	}, nil
}

func (m *mockUserService) RequestEmailChange(ctx context.Context, userID string, req *models.RequestEmailChangeRequest) (*models.EmailChangeResponse, error) {
	if m.requestEmailChangeFunc != nil {
		return m.requestEmailChangeFunc(ctx, userID, req)
	}
	return &models.EmailChangeResponse{
		Message:      "Solicitud de cambio de correo recibida. Pendiente de verificación.",
		PendingEmail: req.GetEmail(),
	}, nil
}

func setupTestRouter(srv services.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	ctrl := controllers.NewAccountController(srv)

	protected := router.Group("/users")
	protected.Use(controllers.AuthRequired())
	{
		protected.GET("/me", ctrl.GetAccount)
		protected.PATCH("/me/username", ctrl.UpdateUsername)
		protected.POST("/me/email", ctrl.RequestEmailChange)
	}

	return router
}

func TestGetAccount_HTTP_Exitoso(t *testing.T) {
	router := setupTestRouter(&mockUserService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
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
		t.Errorf("se esperaba id_usuario 'usr-123', se obtuvo %v", data["id_usuario"])
	}
	if data["username"] != "alanp" {
		t.Errorf("se esperaba username 'alanp', se obtuvo %v", data["username"])
	}
	if data["correo"] != "alan@example.com" {
		t.Errorf("se esperaba correo 'alan@example.com', se obtuvo %v", data["correo"])
	}

	// Verificación de seguridad: nunca exponer contrasena_hash ni password
	if _, exists := data["contrasena_hash"]; exists {
		t.Error("vulnerabilidad de seguridad: contrasena_hash encontrada en la respuesta")
	}
	if _, exists := data["password"]; exists {
		t.Error("vulnerabilidad de seguridad: password encontrada en la respuesta")
	}
}

func TestGetAccount_HTTP_NoAutenticado(t *testing.T) {
	router := setupTestRouter(&mockUserService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	// No se envía encabezado de autenticación

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba código 401 Unauthorized, se obtuvo %d", w.Code)
	}
}

func TestUpdateUsername_HTTP_Exitoso(t *testing.T) {
	router := setupTestRouter(&mockUserService{})

	payload := models.UpdateUsernameRequest{Username: "nuevo_alan"}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/username", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})

	if data["username"] != "nuevo_alan" {
		t.Errorf("se esperaba username 'nuevo_alan', se obtuvo %v", data["username"])
	}
}

func TestUpdateUsername_HTTP_NoAutenticado(t *testing.T) {
	router := setupTestRouter(&mockUserService{})

	payload := models.UpdateUsernameRequest{Username: "nuevo_alan"}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/username", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba código 401 Unauthorized, se obtuvo %d", w.Code)
	}
}

func TestUpdateUsername_HTTP_UsernameDuplicado_Conflicto409(t *testing.T) {
	mockSrv := &mockUserService{
		updateUsernameFunc: func(ctx context.Context, userID string, req *models.UpdateUsernameRequest) (*models.AccountResponse, error) {
			return nil, services.ErrUsernameExists
		},
	}
	router := setupTestRouter(mockSrv)

	payload := models.UpdateUsernameRequest{Username: "duplicado"}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/username", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("se esperaba código 409 Conflict, se obtuvo %d", w.Code)
	}
}

func TestUpdateUsername_HTTP_ValidacionFallida_400(t *testing.T) {
	mockSrv := &mockUserService{
		updateUsernameFunc: func(ctx context.Context, userID string, req *models.UpdateUsernameRequest) (*models.AccountResponse, error) {
			return nil, &services.ValidationError{
				Field:   "username",
				Issue:   "required",
				Message: "El nombre de usuario es obligatorio",
			}
		},
	}
	router := setupTestRouter(mockSrv)

	payload := models.UpdateUsernameRequest{Username: ""}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/username", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba código 400 Bad Request, se obtuvo %d", w.Code)
	}
}

func TestRequestEmailChange_HTTP_Exitoso(t *testing.T) {
	router := setupTestRouter(&mockUserService{})

	payload := models.RequestEmailChangeRequest{Email: "nuevo@example.com"}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/users/me/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})

	if data["pending_email"] != "nuevo@example.com" {
		t.Errorf("se esperaba pending_email 'nuevo@example.com', se obtuvo %v", data["pending_email"])
	}
}

func TestRequestEmailChange_HTTP_CorreoDuplicado_Conflicto409(t *testing.T) {
	mockSrv := &mockUserService{
		requestEmailChangeFunc: func(ctx context.Context, userID string, req *models.RequestEmailChangeRequest) (*models.EmailChangeResponse, error) {
			return nil, services.ErrEmailExists
		},
	}
	router := setupTestRouter(mockSrv)

	payload := models.RequestEmailChangeRequest{Email: "existente@example.com"}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/users/me/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("se esperaba código 409 Conflict, se obtuvo %d", w.Code)
	}
}
