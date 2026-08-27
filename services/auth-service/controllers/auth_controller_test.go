package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/auth-service/controllers"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/services"
)

type mockAuthService struct {
	registerFunc func(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error)
}

func (m *mockAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return &models.RegisterResponse{
		ID:        "mock-uuid-123",
		Username:  req.Username,
		Email:     req.Email,
		CreatedAt: time.Date(2026, 8, 26, 22, 0, 0, 0, time.UTC),
	}, nil
}

func TestRegister_HTTP_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{}
	ctrl := controllers.NewAuthController(srv)
	router.POST("/auth/register", ctrl.Register)

	reqPayload := models.RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("expected success: true, got %v", resp["success"])
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'data' object in response")
	}

	if data["id"] != "mock-uuid-123" {
		t.Errorf("expected id: mock-uuid-123, got %v", data["id"])
	}
	if data["username"] != "newuser" {
		t.Errorf("expected username: newuser, got %v", data["username"])
	}
	if data["email"] != "new@example.com" {
		t.Errorf("expected email: new@example.com, got %v", data["email"])
	}

	// Verify security: password or password hash must never be returned in JSON response
	if _, exists := data["password"]; exists {
		t.Error("security breach: password found in response data")
	}
	if _, exists := data["password_hash"]; exists {
		t.Error("security breach: password_hash found in response data")
	}
}

func TestRegister_HTTP_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{}
	ctrl := controllers.NewAuthController(srv)
	router.POST("/auth/register", ctrl.Register)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["success"] != false {
		t.Errorf("expected success: false, got %v", resp["success"])
	}

	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %v", errObj["code"])
	}
}

func TestRegister_HTTP_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{
		registerFunc: func(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
			return nil, &services.ValidationError{
				Field:   "password",
				Issue:   "too_short",
				Message: "La contraseña debe tener al menos 8 caracteres",
			}
		},
	}
	ctrl := controllers.NewAuthController(srv)
	router.POST("/auth/register", ctrl.Register)

	reqPayload := models.RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "short",
	}
	body, _ := json.Marshal(reqPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", errObj["code"])
	}

	details := errObj["details"].(map[string]interface{})
	if details["field"] != "password" || details["issue"] != "too_short" {
		t.Errorf("expected field:password, issue:too_short. got field:%v, issue:%v", details["field"], details["issue"])
	}
}

func TestRegister_HTTP_DuplicateUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{
		registerFunc: func(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
			return nil, services.ErrUsernameExists
		},
	}
	ctrl := controllers.NewAuthController(srv)
	router.POST("/auth/register", ctrl.Register)

	reqPayload := models.RegisterRequest{
		Username: "existinguser",
		Email:    "new@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "CONFLICT" {
		t.Errorf("expected CONFLICT, got %v", errObj["code"])
	}
}
