package controllers_test

import (
	"bytes"
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
)

type mockAuthService struct {
	registerFunc        func(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error)
	loginFunc           func(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	logoutFunc          func(ctx context.Context, token string) error
	validateSessionFunc func(ctx context.Context, token string) (*models.SessionUser, error)
	changePasswordFunc  func(ctx context.Context, userID string, req *models.ChangePasswordRequest) error
}

func (m *mockAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return &models.RegisterResponse{
		ID:       "mock-uuid-123",
		Username: req.Username,
		Email:    req.Email,
	}, nil
}

func (m *mockAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return &models.LoginResponse{
		SessionToken: "mock-session-token-abc",
		ExpiresIn:    86400,
		User: models.SessionUser{
			ID:       "mock-uuid-123",
			Username: "mockuser",
			Email:    req.Email,
		},
	}, nil
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, token)
	}
	return nil
}

func (m *mockAuthService) ValidateSession(ctx context.Context, token string) (*models.SessionUser, error) {
	if m.validateSessionFunc != nil {
		return m.validateSessionFunc(ctx, token)
	}
	if token == "mock-session-token-abc" || token == "valid-session-token" {
		return &models.SessionUser{
			ID:       "mock-uuid-123",
			Username: "mockuser",
			Email:    "mock@example.com",
		}, nil
	}
	return nil, errors.New("invalid token")
}

func (m *mockAuthService) ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error {
	if m.changePasswordFunc != nil {
		return m.changePasswordFunc(ctx, userID, req)
	}
	return nil
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
	if _, exists := data["contrasena_hash"]; exists {
		t.Error("security breach: contrasena_hash found in response data")
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

	if resp["success"] != false {
		t.Errorf("expected success: false, got %v", resp["success"])
	}

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

func TestLogin_HTTP_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{
		loginFunc: func(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
			return &models.LoginResponse{
				SessionToken: "valid-session-token",
				ExpiresIn:    86400,
				User: models.SessionUser{
					ID:       "user-uuid-1",
					Username: "alan",
					Email:    req.Email,
				},
			}, nil
		},
	}
	ctrl := controllers.NewAuthController(srv)
	router.POST("/auth/login", ctrl.Login)

	reqPayload := models.LoginRequest{
		Email:    "alan@meloop.com",
		Password: "meloop123",
	}
	body, _ := json.Marshal(reqPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("expected success: true, got %v", resp["success"])
	}

	data := resp["data"].(map[string]interface{})
	if data["session_token"] != "valid-session-token" {
		t.Errorf("expected session_token: valid-session-token, got %v", data["session_token"])
	}

	// Verify no sensitive info is returned
	if _, exists := data["password"]; exists {
		t.Error("sensitive info leaked: password found in response")
	}
	if _, exists := data["password_hash"]; exists {
		t.Error("sensitive info leaked: password_hash found in response")
	}
}

func TestLogin_HTTP_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{
		loginFunc: func(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
			return nil, services.ErrInvalidCredentials
		},
	}
	ctrl := controllers.NewAuthController(srv)
	router.POST("/auth/login", ctrl.Login)

	reqPayload := models.LoginRequest{
		Email:    "wrong@meloop.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["success"] != false {
		t.Errorf("expected success: false, got %v", resp["success"])
	}

	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got %v", errObj["code"])
	}
	if errObj["message"] != "Credenciales incorrectas" {
		t.Errorf("expected message 'Credenciales incorrectas', got %v", errObj["message"])
	}
}

func TestLogout_HTTP_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{
		logoutFunc: func(ctx context.Context, token string) error {
			if token != "valid-session-token" {
				return errors.New("invalid token in mock")
			}
			return nil
		},
	}
	ctrl := controllers.NewAuthController(srv)

	// Route protected by middleware
	protected := router.Group("")
	protected.Use(ctrl.AuthRequired())
	protected.POST("/auth/logout", ctrl.Logout)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer valid-session-token")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["success"] != true {
		t.Errorf("expected success: true, got %v", resp["success"])
	}
}

func TestLogout_HTTP_InvalidOrMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{}
	ctrl := controllers.NewAuthController(srv)

	protected := router.Group("")
	protected.Use(ctrl.AuthRequired())
	protected.POST("/auth/logout", ctrl.Logout)

	// Case 1: Missing Token
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/auth/logout", nil)
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing token, got %d", w1.Code)
	}

	// Case 2: Invalid/Expired Token
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/auth/logout", nil)
	req2.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for invalid token, got %d", w2.Code)
	}
}

func TestRegister_HTTP_DuplicateEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{
		registerFunc: func(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
			return nil, services.ErrEmailExists
		},
	}
	ctrl := controllers.NewAuthController(srv)
	router.POST("/auth/register", ctrl.Register)

	reqPayload := models.RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
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

func TestRegister_HTTP_InternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{
		registerFunc: func(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
			return nil, errors.New("unexpected database error")
		},
	}
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

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %v", errObj["code"])
	}
}

func TestValidate_HTTP_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{}
	ctrl := controllers.NewAuthController(srv)

	protected := router.Group("")
	protected.Use(ctrl.AuthRequired())
	protected.GET("/auth/validate", ctrl.Validate)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer valid-session-token")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("expected success: true, got %v", resp["success"])
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	if data["username"] != "mockuser" {
		t.Errorf("expected username: mockuser, got %v", data["username"])
	}
}

func TestChangePassword_HTTP_Exitoso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{}
	ctrl := controllers.NewAuthController(srv)

	protected := router.Group("")
	protected.Use(ctrl.AuthRequired())
	protected.POST("/auth/change-password", ctrl.ChangePassword)

	payload := models.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword456",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-session-token")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error al parsear JSON: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("se esperaba success = true, se obtuvo %v", resp["success"])
	}

	data := resp["data"].(map[string]interface{})
	if data["message"] != "Contraseña actualizada correctamente" {
		t.Errorf("mensaje inesperado: %v", data["message"])
	}
}

func TestChangePassword_HTTP_NoAutenticado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{}
	ctrl := controllers.NewAuthController(srv)

	protected := router.Group("")
	protected.Use(ctrl.AuthRequired())
	protected.POST("/auth/change-password", ctrl.ChangePassword)

	payload := models.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword456",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba código 401 Unauthorized, se obtuvo %d", w.Code)
	}
}

func TestChangePassword_HTTP_ValidacionFallida(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{
		changePasswordFunc: func(ctx context.Context, userID string, req *models.ChangePasswordRequest) error {
			return &services.ValidationError{
				Field:   "new_password",
				Issue:   "too_short",
				Message: "La contraseña debe tener al menos 8 caracteres",
			}
		},
	}
	ctrl := controllers.NewAuthController(srv)

	protected := router.Group("")
	protected.Use(ctrl.AuthRequired())
	protected.POST("/auth/change-password", ctrl.ChangePassword)

	payload := models.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "short",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-session-token")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba código 400 Bad Request, se obtuvo %d", w.Code)
	}
}

func TestChangePassword_HTTP_CredencialesIncorrectas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	srv := &mockAuthService{
		changePasswordFunc: func(ctx context.Context, userID string, req *models.ChangePasswordRequest) error {
			return services.ErrInvalidCredentials
		},
	}
	ctrl := controllers.NewAuthController(srv)

	protected := router.Group("")
	protected.Use(ctrl.AuthRequired())
	protected.POST("/auth/change-password", ctrl.ChangePassword)

	payload := models.ChangePasswordRequest{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword456",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-session-token")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba código 401 Unauthorized, se obtuvo %d", w.Code)
	}
}
