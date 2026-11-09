package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/auth-service/config"
	"github.com/meloop/auth-service/controllers"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/services"
)

// MockUserRepository provee un almacenamiento de usuarios en memoria con aislamiento
type MockUserRepository struct {
	Users              map[string]*models.Usuario
	GetByIDFunc        func(ctx context.Context, id string) (*models.Usuario, error)
	GetByUsernameFunc  func(ctx context.Context, username string) (*models.Usuario, error)
	GetByEmailFunc     func(ctx context.Context, email string) (*models.Usuario, error)
	CreateFunc         func(ctx context.Context, user *models.Usuario) error
	UpdatePasswordFunc func(ctx context.Context, id string, passwordHash string) error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[string]*models.Usuario),
	}
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.Usuario, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	for _, u := range m.Users {
		if u.IDUsuario == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.Usuario, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(ctx, username)
	}
	for _, u := range m.Users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.Usuario, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	for _, u := range m.Users {
		if u.Correo == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.Usuario) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	if user.IDUsuario == "" {
		user.IDUsuario = "mock-uuid-" + user.Username
	}
	m.Users[user.Username] = user
	return nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, id, passwordHash)
	}
	for _, u := range m.Users {
		if u.IDUsuario == id {
			u.ContrasenaHash = passwordHash
			return nil
		}
	}
	return nil
}

// MockSessionRepository provee almacenamiento temporal de sesiones en memoria
type MockSessionRepository struct {
	Sessions           map[string]*models.SessionUser
	CreateFunc         func(ctx context.Context, token string, user *models.SessionUser, ttl time.Duration) error
	GetFunc            func(ctx context.Context, token string) (*models.SessionUser, error)
	DeleteFunc         func(ctx context.Context, token string) error
	DeleteByUserIDFunc func(ctx context.Context, userID string) error
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		Sessions: make(map[string]*models.SessionUser),
	}
}

func (m *MockSessionRepository) Create(ctx context.Context, token string, user *models.SessionUser, ttl time.Duration) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, token, user, ttl)
	}
	m.Sessions[token] = user
	return nil
}

func (m *MockSessionRepository) Get(ctx context.Context, token string) (*models.SessionUser, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, token)
	}
	return m.Sessions[token], nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, token string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, token)
	}
	delete(m.Sessions, token)
	return nil
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if m.DeleteByUserIDFunc != nil {
		return m.DeleteByUserIDFunc(ctx, userID)
	}
	for k, v := range m.Sessions {
		if v != nil && v.ID == userID {
			delete(m.Sessions, k)
		}
	}
	return nil
}

// MockPasswordRecoveryRepository provee almacenamiento en memoria para tokens de recuperación
type MockPasswordRecoveryRepository struct {
	Tokens              map[string]*models.PasswordRecoveryToken
	CreateTokenFunc     func(ctx context.Context, token *models.PasswordRecoveryToken) error
	GetTokenByHashFunc  func(ctx context.Context, tokenHash string) (*models.PasswordRecoveryToken, error)
	MarkTokenAsUsedFunc func(ctx context.Context, idRecuperacion string) error
}

func NewMockPasswordRecoveryRepository() *MockPasswordRecoveryRepository {
	return &MockPasswordRecoveryRepository{
		Tokens: make(map[string]*models.PasswordRecoveryToken),
	}
}

func (m *MockPasswordRecoveryRepository) CreateToken(ctx context.Context, token *models.PasswordRecoveryToken) error {
	if m.CreateTokenFunc != nil {
		return m.CreateTokenFunc(ctx, token)
	}
	m.Tokens[token.TokenHash] = token
	return nil
}

func (m *MockPasswordRecoveryRepository) GetTokenByHash(ctx context.Context, tokenHash string) (*models.PasswordRecoveryToken, error) {
	if m.GetTokenByHashFunc != nil {
		return m.GetTokenByHashFunc(ctx, tokenHash)
	}
	return m.Tokens[tokenHash], nil
}

func (m *MockPasswordRecoveryRepository) MarkTokenAsUsed(ctx context.Context, idRecuperacion string) error {
	if m.MarkTokenAsUsedFunc != nil {
		return m.MarkTokenAsUsedFunc(ctx, idRecuperacion)
	}
	for _, t := range m.Tokens {
		if t.IDRecuperacion == idRecuperacion {
			t.Usado = true
			now := time.Now()
			t.UsadoEn = &now
			return nil
		}
	}
	return nil
}

// MockEmailService captura correos enviados
type MockEmailService struct {
	SentEmails               []string
	SentLinks                []string
	SendPasswordRecoveryFunc func(ctx context.Context, toEmail string, recoveryLink string) error
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (m *MockEmailService) SendPasswordRecoveryEmail(ctx context.Context, toEmail string, recoveryLink string) error {
	if m.SendPasswordRecoveryFunc != nil {
		return m.SendPasswordRecoveryFunc(ctx, toEmail, recoveryLink)
	}
	m.SentEmails = append(m.SentEmails, toEmail)
	m.SentLinks = append(m.SentLinks, recoveryLink)
	return nil
}

// TestApp encapsula una instancia aislada del servicio y su router Gin
type TestApp struct {
	Config       *config.Config
	UserRepo     *MockUserRepository
	SessionRepo  *MockSessionRepository
	RecoveryRepo *MockPasswordRecoveryRepository
	EmailService *MockEmailService
	AuthService  services.AuthService
	Controller   *controllers.AuthController
	Router       *gin.Engine
}

func setupTestApp(t *testing.T) *TestApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		PasswordMinLength: 8,
		SessionTTL:        24 * time.Hour,
		RecoveryTokenTTL:  30 * time.Minute,
		RecoveryURLBase:   "http://localhost:8080/auth/password-recovery/reset?token=",
	}

	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	recoveryRepo := NewMockPasswordRecoveryRepository()
	emailService := NewMockEmailService()

	authSrv := services.NewAuthService(cfg, userRepo, sessionRepo, recoveryRepo, emailService)
	authCtrl := controllers.NewAuthController(authSrv)

	router := gin.New()
	router.POST("/auth/register", authCtrl.Register)
	router.POST("/auth/login", authCtrl.Login)
	router.POST("/auth/password-recovery", authCtrl.RequestPasswordRecovery)
	router.POST("/auth/password-recovery/reset", authCtrl.ResetPassword)

	protected := router.Group("")
	protected.Use(authCtrl.AuthRequired())
	{
		protected.POST("/auth/logout", authCtrl.Logout)
		protected.GET("/auth/validate", authCtrl.Validate)
		protected.POST("/auth/change-password", authCtrl.ChangePassword)
	}

	return &TestApp{
		Config:       cfg,
		UserRepo:     userRepo,
		SessionRepo:  sessionRepo,
		RecoveryRepo: recoveryRepo,
		EmailService: emailService,
		AuthService:  authSrv,
		Controller:   authCtrl,
		Router:       router,
	}
}

// performRequest ejecuta una petición HTTP sobre el router Gin y deserializa la respuesta
func performRequest(router *gin.Engine, method, path string, body interface{}, headers map[string]string) (*httptest.ResponseRecorder, map[string]interface{}) {
	var bodyBuf *bytes.Buffer
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyBuf = bytes.NewBufferString(v)
		case []byte:
			bodyBuf = bytes.NewBuffer(v)
		default:
			data, _ := json.Marshal(body)
			bodyBuf = bytes.NewBuffer(data)
		}
	} else {
		bodyBuf = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, bodyBuf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	if w.Body.Len() > 0 {
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
	}

	return w, resp
}
