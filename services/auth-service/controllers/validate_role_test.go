package controllers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/auth-service/controllers"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/services"
	"github.com/meloop/services/common/httpresponse"
)

func newValidateRouter(srv services.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NewAuthController(srv)

	protected := router.Group("")
	protected.Use(ctrl.AuthRequired())
	protected.GET("/auth/validate", ctrl.Validate)
	return router
}

func validateRequest(t *testing.T, router *gin.Engine, token string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/validate", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeValidateResponse(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v; body = %s", err, recorder.Body.String())
	}
	return response
}

func assertValidateSuccess(t *testing.T, recorder *httptest.ResponseRecorder, wantRole string) {
	t.Helper()
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", recorder.Code, recorder.Body.String())
	}

	response := decodeValidateResponse(t, recorder)
	if response["success"] != true {
		t.Fatalf("success = %v, want true", response["success"])
	}
	data, ok := response["data"].(map[string]any)
	if !ok {
		t.Fatalf("data no es objeto: %v", response["data"])
	}
	if data["id"] != "mock-uuid-123" {
		t.Errorf("id = %v, want mock-uuid-123", data["id"])
	}
	if data["username"] != "mockuser" {
		t.Errorf("username = %v, want mockuser", data["username"])
	}
	if data["email"] != "mock@example.com" {
		t.Errorf("email = %v, want mock@example.com", data["email"])
	}
	if data["role"] != wantRole {
		t.Errorf("role = %v, want %q", data["role"], wantRole)
	}
	if _, exists := data["password"]; exists {
		t.Error("la respuesta expone password")
	}
}

func TestValidateReturnsRoleAdditively(t *testing.T) {
	srv := &mockAuthService{
		getRoleByIDFunc: func(ctx context.Context, userID string) (string, error) {
			if userID != "mock-uuid-123" {
				t.Errorf("userID = %q, want mock-uuid-123", userID)
			}
			return "admin", nil
		},
	}
	recorder := validateRequest(t, newValidateRouter(srv), "valid-session-token")
	assertValidateSuccess(t, recorder, "admin")
}

func TestValidateInvalidSessionReturns401WithoutQueryingRole(t *testing.T) {
	queried := false
	srv := &mockAuthService{
		validateSessionFunc: func(ctx context.Context, token string) (*models.SessionUser, error) {
			return nil, services.ErrInvalidSession
		},
		getRoleByIDFunc: func(ctx context.Context, userID string) (string, error) {
			queried = true
			return "admin", nil
		},
	}
	recorder := validateRequest(t, newValidateRouter(srv), "expired-token")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body = %s", recorder.Code, recorder.Body.String())
	}
	response := decodeValidateResponse(t, recorder)
	errBody, ok := response["error"].(map[string]any)
	if !ok || errBody["code"] != httpresponse.ErrUnauthorized {
		t.Fatalf("error inesperado: %v", response["error"])
	}
	if queried {
		t.Fatal("no se debió consultar el rol para una sesión inválida")
	}
}

func TestValidateReflectsRoleChangeDuringActiveSession(t *testing.T) {
	currentRole := "moderador"
	srv := &mockAuthService{
		getRoleByIDFunc: func(ctx context.Context, userID string) (string, error) {
			return currentRole, nil
		},
	}
	router := newValidateRouter(srv)

	assertValidateSuccess(t, validateRequest(t, router, "valid-session-token"), "moderador")

	currentRole = "admin"
	assertValidateSuccess(t, validateRequest(t, router, "valid-session-token"), "admin")
}

func TestValidateRoleQueryErrorReturns503(t *testing.T) {
	srv := &mockAuthService{
		getRoleByIDFunc: func(ctx context.Context, userID string) (string, error) {
			return "", errors.New("postgres caído")
		},
	}
	recorder := validateRequest(t, newValidateRouter(srv), "valid-session-token")

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body = %s", recorder.Code, recorder.Body.String())
	}
	response := decodeValidateResponse(t, recorder)
	errBody, ok := response["error"].(map[string]any)
	if !ok || errBody["code"] != "DATABASE_UNAVAILABLE" {
		t.Fatalf("error inesperado: %v", response["error"])
	}
	if response["data"] != nil {
		t.Fatalf("no se debió devolver data ante error de consulta: %v", response["data"])
	}
}

func TestValidateUserNoLongerExistsReturns401(t *testing.T) {
	srv := &mockAuthService{
		getRoleByIDFunc: func(ctx context.Context, userID string) (string, error) {
			return "", services.ErrUserNotFound
		},
	}
	recorder := validateRequest(t, newValidateRouter(srv), "valid-session-token")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body = %s", recorder.Code, recorder.Body.String())
	}
	response := decodeValidateResponse(t, recorder)
	errBody, ok := response["error"].(map[string]any)
	if !ok || errBody["code"] != httpresponse.ErrUnauthorized {
		t.Fatalf("error inesperado: %v", response["error"])
	}
	if response["data"] != nil {
		t.Fatalf("no se debió devolver data: %v", response["data"])
	}
}

func TestValidateUnknownRoleGrantsNothing(t *testing.T) {
	srv := &mockAuthService{
		getRoleByIDFunc: func(ctx context.Context, userID string) (string, error) {
			return "root", nil
		},
	}
	recorder := validateRequest(t, newValidateRouter(srv), "valid-session-token")
	assertValidateSuccess(t, recorder, "")
}
