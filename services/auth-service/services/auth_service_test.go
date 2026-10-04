package services_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/meloop/auth-service/config"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/repositories"
	"github.com/meloop/auth-service/services"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	users              map[string]*models.Usuario
	getByIDFunc        func(ctx context.Context, id string) (*models.Usuario, error)
	getByUsernameFunc  func(ctx context.Context, username string) (*models.Usuario, error)
	getByEmailFunc     func(ctx context.Context, email string) (*models.Usuario, error)
	createFunc         func(ctx context.Context, user *models.Usuario) error
	updatePasswordFunc func(ctx context.Context, id string, passwordHash string) error
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*models.Usuario, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	for _, u := range m.users {
		if u.IDUsuario == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*models.Usuario, error) {
	if m.getByUsernameFunc != nil {
		return m.getByUsernameFunc(ctx, username)
	}
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.Usuario, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	for _, u := range m.users {
		if u.Correo == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.Usuario) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	user.IDUsuario = "mock-usuario-uuid-123"
	m.users[user.Username] = user
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	if m.updatePasswordFunc != nil {
		return m.updatePasswordFunc(ctx, id, passwordHash)
	}
	for _, u := range m.users {
		if u.IDUsuario == id {
			u.ContrasenaHash = passwordHash
			return nil
		}
	}
	return nil
}

type mockSessionRepository struct {
	sessions           map[string]*models.SessionUser
	createFunc         func(ctx context.Context, token string, user *models.SessionUser, ttl time.Duration) error
	getFunc            func(ctx context.Context, token string) (*models.SessionUser, error)
	deleteFunc         func(ctx context.Context, token string) error
	deleteByUserIDFunc func(ctx context.Context, userID string) error
}

func (m *mockSessionRepository) Create(ctx context.Context, token string, user *models.SessionUser, ttl time.Duration) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, token, user, ttl)
	}
	m.sessions[token] = user
	return nil
}

func (m *mockSessionRepository) Get(ctx context.Context, token string) (*models.SessionUser, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, token)
	}
	return m.sessions[token], nil
}

func (m *mockSessionRepository) Delete(ctx context.Context, token string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, token)
	}
	delete(m.sessions, token)
	return nil
}

func (m *mockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if m.deleteByUserIDFunc != nil {
		return m.deleteByUserIDFunc(ctx, userID)
	}
	for k, v := range m.sessions {
		if v != nil && v.ID == userID {
			delete(m.sessions, k)
		}
	}
	return nil
}

type mockPasswordRecoveryRepository struct {
	tokens              map[string]*models.PasswordRecoveryToken
	createTokenFunc     func(ctx context.Context, token *models.PasswordRecoveryToken) error
	getTokenByHashFunc  func(ctx context.Context, tokenHash string) (*models.PasswordRecoveryToken, error)
	markTokenAsUsedFunc func(ctx context.Context, idRecuperacion string) error
}

func (m *mockPasswordRecoveryRepository) CreateToken(ctx context.Context, token *models.PasswordRecoveryToken) error {
	if m.createTokenFunc != nil {
		return m.createTokenFunc(ctx, token)
	}
	m.tokens[token.TokenHash] = token
	return nil
}

func (m *mockPasswordRecoveryRepository) GetTokenByHash(ctx context.Context, tokenHash string) (*models.PasswordRecoveryToken, error) {
	if m.getTokenByHashFunc != nil {
		return m.getTokenByHashFunc(ctx, tokenHash)
	}
	t, ok := m.tokens[tokenHash]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (m *mockPasswordRecoveryRepository) MarkTokenAsUsed(ctx context.Context, idRecuperacion string) error {
	if m.markTokenAsUsedFunc != nil {
		return m.markTokenAsUsedFunc(ctx, idRecuperacion)
	}
	for _, t := range m.tokens {
		if t.IDRecuperacion == idRecuperacion {
			t.Usado = true
			now := time.Now()
			t.UsadoEn = &now
			return nil
		}
	}
	return nil
}

type mockEmailService struct {
	sentEmails               []string
	sentLinks                []string
	sendPasswordRecoveryFunc func(ctx context.Context, toEmail string, recoveryLink string) error
}

func (m *mockEmailService) SendPasswordRecoveryEmail(ctx context.Context, toEmail string, recoveryLink string) error {
	if m.sendPasswordRecoveryFunc != nil {
		return m.sendPasswordRecoveryFunc(ctx, toEmail, recoveryLink)
	}
	m.sentEmails = append(m.sentEmails, toEmail)
	m.sentLinks = append(m.sentLinks, recoveryLink)
	return nil
}

func createTestAuthService(
	cfg *config.Config,
	repo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	recoveryRepo repositories.PasswordRecoveryRepository,
	emailService services.EmailService,
) services.AuthService {
	if recoveryRepo == nil {
		recoveryRepo = &mockPasswordRecoveryRepository{tokens: make(map[string]*models.PasswordRecoveryToken)}
	}
	if emailService == nil {
		emailService = &mockEmailService{}
	}
	return services.NewAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)
}

func TestRegister_Success(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockUserRepository{users: make(map[string]*models.Usuario)}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	req := &models.RegisterRequest{
		Username: "validuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	res, err := srv.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.ID != "mock-usuario-uuid-123" {
		t.Errorf("expected ID 'mock-usuario-uuid-123', got %s", res.ID)
	}
	if res.Username != req.Username {
		t.Errorf("expected username %s, got %s", req.Username, res.Username)
	}
	if res.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, res.Email)
	}

	savedUser := repo.users[req.Username]
	if savedUser == nil {
		t.Fatal("expected user to be persisted in repo")
	}
	if savedUser.IDUsuario != "mock-usuario-uuid-123" {
		t.Errorf("expected IDUsuario 'mock-usuario-uuid-123', got %s", savedUser.IDUsuario)
	}
	if savedUser.Username != req.Username {
		t.Errorf("expected username %s, got %s", req.Username, savedUser.Username)
	}
	if savedUser.Correo != req.Email {
		t.Errorf("expected correo %s, got %s", req.Email, savedUser.Correo)
	}
	if savedUser.Experiencia != 0 {
		t.Errorf("expected initial experiencia to be 0, got %d", savedUser.Experiencia)
	}
	if savedUser.IDNivel != nil {
		t.Errorf("expected default IDNivel to be nil when not configured, got %v", savedUser.IDNivel)
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockUserRepository{users: make(map[string]*models.Usuario)}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	tests := []struct {
		name          string
		req           *models.RegisterRequest
		expectedField string
		expectedIssue string
	}{
		{
			name: "empty username",
			req: &models.RegisterRequest{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedField: "username",
			expectedIssue: "required",
		},
		{
			name: "username too long",
			req: &models.RegisterRequest{
				Username: "thisusernameiswaytoolongexceedingthelimitoffiftycharacters",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedField: "username",
			expectedIssue: "too_long",
		},
		{
			name: "empty email",
			req: &models.RegisterRequest{
				Username: "validuser",
				Email:    "",
				Password: "password123",
			},
			expectedField: "email",
			expectedIssue: "required",
		},
		{
			name: "email too long",
			req: &models.RegisterRequest{
				Username: "validuser",
				Email:    "thisemailaddressiswaytoolongexceedingthelimitof255characterswhichisextremelyunusualbutmustbecheckeddefensivelytoavoidbufferissuesordatabaseproblemsathandbecauseintegritymattersandwearegoingtomakeitlongerbyaddingmoreandmoretextuntilitexceeds255charactersdefinitively@example.com",
				Password: "password123",
			},
			expectedField: "email",
			expectedIssue: "too_long",
		},
		{
			name: "invalid email format",
			req: &models.RegisterRequest{
				Username: "validuser",
				Email:    "invalid-email-address",
				Password: "password123",
			},
			expectedField: "email",
			expectedIssue: "invalid_format",
		},
		{
			name: "empty password",
			req: &models.RegisterRequest{
				Username: "validuser",
				Email:    "test@example.com",
				Password: "",
			},
			expectedField: "password",
			expectedIssue: "required",
		},
		{
			name: "password too long",
			req: &models.RegisterRequest{
				Username: "validuser",
				Email:    "test@example.com",
				Password: strings.Repeat("a", 73),
			},
			expectedField: "password",
			expectedIssue: "too_long",
		},
		{
			name: "password too short",
			req: &models.RegisterRequest{
				Username: "validuser",
				Email:    "test@example.com",
				Password: "short",
			},
			expectedField: "password",
			expectedIssue: "too_short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := srv.Register(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			var valErr *services.ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected ValidationError, got %T", err)
			}

			if valErr.Field != tt.expectedField {
				t.Errorf("expected field %s, got %s", tt.expectedField, valErr.Field)
			}
			if valErr.Issue != tt.expectedIssue {
				t.Errorf("expected issue %s, got %s", tt.expectedIssue, valErr.Issue)
			}
		})
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{
			"existinguser": {
				IDUsuario:      "id-1",
				Username:       "existinguser",
				Correo:         "first@example.com",
				ContrasenaHash: "hash1",
			},
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	req := &models.RegisterRequest{
		Username: "existinguser",
		Email:    "second@example.com",
		Password: "password123",
	}
	_, err := srv.Register(context.Background(), req)
	if !errors.Is(err, services.ErrUsernameExists) {
		t.Errorf("expected ErrUsernameExists, got %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{
			"firstuser": {
				IDUsuario:      "id-1",
				Username:       "firstuser",
				Correo:         "existing@example.com",
				ContrasenaHash: "hash1",
			},
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	req := &models.RegisterRequest{
		Username: "seconduser",
		Email:    "existing@example.com",
		Password: "password123",
	}
	_, err := srv.Register(context.Background(), req)
	if !errors.Is(err, services.ErrEmailExists) {
		t.Errorf("expected ErrEmailExists, got %v", err)
	}
}

func TestRegister_PersistenceError(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	expectedErr := errors.New("database disk I/O error")
	repo := &mockUserRepository{
		users: make(map[string]*models.Usuario),
		createFunc: func(ctx context.Context, user *models.Usuario) error {
			return expectedErr
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	req := &models.RegisterRequest{
		Username: "newuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err := srv.Register(context.Background(), req)
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected persistence error %v, got %v", expectedErr, err)
	}
}

func TestRegister_BCryptHashVerification(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockUserRepository{users: make(map[string]*models.Usuario)}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	rawPassword := "SuperSecretPassword2026!"
	req := &models.RegisterRequest{
		Username: "secureuser",
		Email:    "secure@example.com",
		Password: rawPassword,
	}

	_, err := srv.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected registration to succeed, got %v", err)
	}

	savedUser := repo.users[req.Username]
	if savedUser == nil {
		t.Fatal("expected user to be in repository")
	}

	if savedUser.ContrasenaHash == rawPassword {
		t.Fatal("CRITICAL: Raw plaintext password was saved in contrasena_hash!")
	}

	if !strings.HasPrefix(savedUser.ContrasenaHash, "$2a$") && !strings.HasPrefix(savedUser.ContrasenaHash, "$2b$") {
		t.Errorf("expected contrasena_hash to start with bcrypt prefix ($2a$ or $2b$), got %s", savedUser.ContrasenaHash)
	}

	err = bcrypt.CompareHashAndPassword([]byte(savedUser.ContrasenaHash), []byte(rawPassword))
	if err != nil {
		t.Errorf("bcrypt verification failed: %v", err)
	}
}

func TestRegister_DatabaseUniqueViolationRace(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}

	repoUsernameRace := &mockUserRepository{
		users: make(map[string]*models.Usuario),
		createFunc: func(ctx context.Context, user *models.Usuario) error {
			return errors.New("ERROR: duplicate key value violates unique constraint \"usuario_username_key\" (SQLSTATE 23505)")
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srvUsername := createTestAuthService(cfg, repoUsernameRace, sessionRepo, nil, nil)
	req1 := &models.RegisterRequest{
		Username: "duplicateuser",
		Email:    "new@example.com",
		Password: "password123",
	}
	_, err := srvUsername.Register(context.Background(), req1)
	if !errors.Is(err, services.ErrUsernameExists) {
		t.Errorf("expected ErrUsernameExists from database race mapping, got %v", err)
	}

	repoEmailRace := &mockUserRepository{
		users: make(map[string]*models.Usuario),
		createFunc: func(ctx context.Context, user *models.Usuario) error {
			return errors.New("ERROR: duplicate key value violates unique constraint \"usuario_correo_key\" (SQLSTATE 23505)")
		},
	}
	srvEmail := createTestAuthService(cfg, repoEmailRace, sessionRepo, nil, nil)
	req2 := &models.RegisterRequest{
		Username: "newuser2",
		Email:    "duplicate@example.com",
		Password: "password123",
	}
	_, err = srvEmail.Register(context.Background(), req2)
	if !errors.Is(err, services.ErrEmailExists) {
		t.Errorf("expected ErrEmailExists from database race mapping, got %v", err)
	}
}

// RF-02 Login: 1. Login con usuario existente y contraseña correcta
// RF-02 Login: 4. Creación de sesión después de login exitoso
func TestLogin_Success(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("meloop123"), bcrypt.DefaultCost)
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{
			"alan": {
				IDUsuario:      "user-uuid-1",
				Username:       "alan",
				Correo:         "alan@meloop.com",
				ContrasenaHash: string(hashedPassword),
			},
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	req := &models.LoginRequest{
		Email:    "alan@meloop.com",
		Password: "meloop123",
	}

	res, err := srv.Login(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.SessionToken == "" {
		t.Error("expected session token to be generated")
	}
	if res.ExpiresIn != 86400 {
		t.Errorf("expected expires_in to be 86400, got %d", res.ExpiresIn)
	}
	if res.User.ID != "user-uuid-1" || res.User.Username != "alan" || res.User.Email != "alan@meloop.com" {
		t.Errorf("unexpected user details returned: %+v", res.User)
	}

	// Verify it was saved in session repository
	savedSession := sessionRepo.sessions[res.SessionToken]
	if savedSession == nil {
		t.Fatal("expected session to be saved in sessionRepo")
	}
	if savedSession.ID != "user-uuid-1" {
		t.Errorf("expected saved session user ID to be user-uuid-1, got %s", savedSession.ID)
	}
}

// RF-02 Login: 2. Login con contraseña incorrecta
// RF-02 Login: 3. Login con usuario inexistente
func TestLogin_InvalidCredentials(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("meloop123"), bcrypt.DefaultCost)
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{
			"alan": {
				IDUsuario:      "user-uuid-1",
				Username:       "alan",
				Correo:         "alan@meloop.com",
				ContrasenaHash: string(hashedPassword),
			},
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	// Case 1: Usuario inexistente (email no encontrado)
	req1 := &models.LoginRequest{
		Email:    "wrong@meloop.com",
		Password: "meloop123",
	}
	_, err := srv.Login(context.Background(), req1)
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for non-existent email, got %v", err)
	}

	// Case 2: Contraseña incorrecta
	req2 := &models.LoginRequest{
		Email:    "alan@meloop.com",
		Password: "wrongpassword",
	}
	_, err = srv.Login(context.Background(), req2)
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for incorrect password, got %v", err)
	}
}

// RF-03 Logout: 5. Logout
func TestLogout_Success(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	repo := &mockUserRepository{users: make(map[string]*models.Usuario)}
	sessionRepo := &mockSessionRepository{
		sessions: map[string]*models.SessionUser{
			"valid-token-123": {
				ID:       "user-uuid-1",
				Username: "alan",
				Email:    "alan@meloop.com",
			},
		},
	}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	err := srv.Logout(context.Background(), "valid-token-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Session should be deleted from session repository
	if _, exists := sessionRepo.sessions["valid-token-123"]; exists {
		t.Error("expected session to be deleted from repo")
	}
}

// RF-03 / Middleware: 6. Rechazo de una sesión después de logout
// RF-03 / Middleware: 7. Rechazo de una sesión expirada/inexistente
func TestValidateSession(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	repo := &mockUserRepository{users: map[string]*models.Usuario{
		"alan": {IDUsuario: "user-uuid-1", Username: "alan", Correo: "alan@meloop.com"},
	}}
	sessionRepo := &mockSessionRepository{
		sessions: map[string]*models.SessionUser{
			"valid-token-123": {
				ID:       "user-uuid-1",
				Username: "alan",
				Email:    "alan@meloop.com",
			},
		},
	}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	// Case 1: Valid token
	user, err := srv.ValidateSession(context.Background(), "valid-token-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil || user.ID != "user-uuid-1" {
		t.Errorf("unexpected user returned: %+v", user)
	}

	// Case 2: Invalidate session with Logout, then test rejection (Scenario 6)
	err = srv.Logout(context.Background(), "valid-token-123")
	if err != nil {
		t.Fatalf("expected successful logout, got %v", err)
	}

	_, err = srv.ValidateSession(context.Background(), "valid-token-123")
	if err == nil {
		t.Error("expected error when validating session after logout, got nil")
	}

	// Case 3: Non-existent / expired token (Scenario 7)
	_, err = srv.ValidateSession(context.Background(), "never-existed-token")
	if err == nil {
		t.Error("expected error for non-existent session token, got nil")
	}
}

func TestValidateSessionRejectsAndDeletesSuspendedAccount(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	repo := &mockUserRepository{users: map[string]*models.Usuario{
		"suspended": {IDUsuario: "user-uuid-suspended", Suspendido: true},
	}}
	sessionRepo := &mockSessionRepository{sessions: map[string]*models.SessionUser{
		"suspended-token": {ID: "user-uuid-suspended"},
	}}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	if _, err := srv.ValidateSession(context.Background(), "suspended-token"); !errors.Is(err, services.ErrInvalidCredentials) {
		t.Fatalf("ValidateSession() error = %v, want ErrInvalidCredentials", err)
	}
	if _, exists := sessionRepo.sessions["suspended-token"]; exists {
		t.Fatal("expected suspended session to be deleted")
	}
}

func TestLoginRejectsSuspendedAccount(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("meloop123"), bcrypt.DefaultCost)
	repo := &mockUserRepository{users: map[string]*models.Usuario{
		"suspended": {
			IDUsuario:      "user-uuid-suspended",
			Username:       "suspended",
			Correo:         "suspended@example.com",
			ContrasenaHash: string(hashedPassword),
			Suspendido:     true,
		},
	}}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	_, err := srv.Login(context.Background(), &models.LoginRequest{Email: "suspended@example.com", Password: "meloop123"})
	if !errors.Is(err, services.ErrAccountSuspended) {
		t.Fatalf("Login() error = %v, want ErrAccountSuspended", err)
	}
}

func TestChangePassword_Exitoso(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	initialHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.Usuario{
		IDUsuario:      "user-uuid-123",
		Username:       "alan",
		Correo:         "alan@example.com",
		ContrasenaHash: string(initialHash),
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"alan": user},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	req := &models.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newPassword456",
	}

	err := srv.ChangePassword(context.Background(), "user-uuid-123", req)
	if err != nil {
		t.Fatalf("se esperaba cambio de contraseña exitoso, se obtuvo: %v", err)
	}

	// Verificar que el nuevo hash coincida con la nueva contraseña
	err = bcrypt.CompareHashAndPassword([]byte(user.ContrasenaHash), []byte("newPassword456"))
	if err != nil {
		t.Errorf("el hash persistido no coincide con la nueva contraseña: %v", err)
	}
}

func TestChangePassword_Validaciones(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	initialHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.Usuario{
		IDUsuario:      "user-uuid-123",
		Username:       "alan",
		Correo:         "alan@example.com",
		ContrasenaHash: string(initialHash),
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"alan": user},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	casos := []struct {
		nombre        string
		req           *models.ChangePasswordRequest
		expectedField string
		expectedIssue string
	}{
		{
			nombre: "contraseña actual vacía",
			req: &models.ChangePasswordRequest{
				CurrentPassword: "",
				NewPassword:     "newPassword456",
			},
			expectedField: "current_password",
			expectedIssue: "required",
		},
		{
			nombre: "nueva contraseña vacía",
			req: &models.ChangePasswordRequest{
				CurrentPassword: "password123",
				NewPassword:     "",
			},
			expectedField: "new_password",
			expectedIssue: "required",
		},
		{
			nombre: "nueva contraseña menor al mínimo",
			req: &models.ChangePasswordRequest{
				CurrentPassword: "password123",
				NewPassword:     "corta",
			},
			expectedField: "new_password",
			expectedIssue: "too_short",
		},
		{
			nombre: "nueva contraseña mayor a 72 caracteres",
			req: &models.ChangePasswordRequest{
				CurrentPassword: "password123",
				NewPassword:     strings.Repeat("a", 73),
			},
			expectedField: "new_password",
			expectedIssue: "too_long",
		},
		{
			nombre: "nueva contraseña idéntica a la actual",
			req: &models.ChangePasswordRequest{
				CurrentPassword: "password123",
				NewPassword:     "password123",
			},
			expectedField: "new_password",
			expectedIssue: "same_as_current",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := srv.ChangePassword(context.Background(), "user-uuid-123", c.req)
			if err == nil {
				t.Fatal("se esperaba error de validación, pero fue nil")
			}
			valErr, ok := err.(*services.ValidationError)
			if !ok {
				t.Fatalf("se esperaba *ValidationError, se obtuvo %T", err)
			}
			if valErr.Field != c.expectedField {
				t.Errorf("campo esperado '%s', obtenido '%s'", c.expectedField, valErr.Field)
			}
			if valErr.Issue != c.expectedIssue {
				t.Errorf("issue esperado '%s', obtenido '%s'", c.expectedIssue, valErr.Issue)
			}
		})
	}
}

func TestChangePassword_ContrasenaActualIncorrecta(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	initialHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.Usuario{
		IDUsuario:      "user-uuid-123",
		Username:       "alan",
		Correo:         "alan@example.com",
		ContrasenaHash: string(initialHash),
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"alan": user},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	req := &models.ChangePasswordRequest{
		CurrentPassword: "contrasena_erronea",
		NewPassword:     "newPassword456",
	}

	err := srv.ChangePassword(context.Background(), "user-uuid-123", req)
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials, se obtuvo: %v", err)
	}
}

func TestChangePassword_UsuarioNoExiste(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockUserRepository{users: make(map[string]*models.Usuario)}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := createTestAuthService(cfg, repo, sessionRepo, nil, nil)

	req := &models.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newPassword456",
	}

	err := srv.ChangePassword(context.Background(), "user-inexistente", req)
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials cuando el usuario no existe, se obtuvo: %v", err)
	}
}

// =============================================================================
// PRUEBAS RF-04: Recuperación de acceso mediante correo
// =============================================================================

// 1. Solicitud con correo válido
func TestRequestPasswordRecovery_ValidEmail(t *testing.T) {
	cfg := &config.Config{
		RecoveryTokenTTL: 30 * time.Minute,
		RecoveryURLBase:  "http://localhost:8080/auth/password-recovery/reset?token=",
	}
	user := &models.Usuario{
		IDUsuario: "user-uuid-recovery-1",
		Username:  "alan",
		Correo:    "alan@meloop.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"alan": user},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	recoveryRepo := &mockPasswordRecoveryRepository{tokens: make(map[string]*models.PasswordRecoveryToken)}
	emailService := &mockEmailService{}

	srv := createTestAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)

	err := srv.RequestPasswordRecovery(context.Background(), &models.PasswordRecoveryRequest{
		Email: "alan@meloop.com",
	})
	if err != nil {
		t.Fatalf("expected no error for valid email recovery request, got: %v", err)
	}

	// Verificar que se envió el correo
	if len(emailService.sentEmails) != 1 || emailService.sentEmails[0] != "alan@meloop.com" {
		t.Fatalf("expected email to be sent to alan@meloop.com, got: %v", emailService.sentEmails)
	}
	if len(emailService.sentLinks) != 1 || !strings.Contains(emailService.sentLinks[0], "token=") {
		t.Fatalf("expected recovery link in email, got: %v", emailService.sentLinks)
	}

	// Verificar que el token se haya persistido en hash
	if len(recoveryRepo.tokens) != 1 {
		t.Fatalf("expected 1 recovery token in repository, got: %d", len(recoveryRepo.tokens))
	}
	for hash, tokenRecord := range recoveryRepo.tokens {
		if tokenRecord.IDUsuario != "user-uuid-recovery-1" {
			t.Errorf("expected token for user-uuid-recovery-1, got: %s", tokenRecord.IDUsuario)
		}
		if tokenRecord.Usado {
			t.Errorf("expected token not to be used yet")
		}
		if len(hash) != 64 {
			t.Errorf("expected 64-char sha256 token hash, got length %d", len(hash))
		}
	}
}

// 2. Solicitud con correo inexistente sin revelar su existencia
func TestRequestPasswordRecovery_NonExistentEmail_NoLeak(t *testing.T) {
	cfg := &config.Config{
		RecoveryTokenTTL: 30 * time.Minute,
		RecoveryURLBase:  "http://localhost:8080/auth/password-recovery/reset?token=",
	}
	repo := &mockUserRepository{users: make(map[string]*models.Usuario)}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	recoveryRepo := &mockPasswordRecoveryRepository{tokens: make(map[string]*models.PasswordRecoveryToken)}
	emailService := &mockEmailService{}

	srv := createTestAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)

	// Correo inexistente no debe retornar error ni enviar correos ni persistir tokens
	err := srv.RequestPasswordRecovery(context.Background(), &models.PasswordRecoveryRequest{
		Email: "nonexistent@meloop.com",
	})
	if err != nil {
		t.Fatalf("expected nil error for non-existent email (security no-leak requirement), got: %v", err)
	}

	if len(emailService.sentEmails) != 0 {
		t.Errorf("expected no email to be sent for non-existent user, got %d", len(emailService.sentEmails))
	}
	if len(recoveryRepo.tokens) != 0 {
		t.Errorf("expected no token saved for non-existent user, got %d", len(recoveryRepo.tokens))
	}
}

// 3. Generación y hash de token seguro (expiración a 30 minutos)
func TestRequestPasswordRecovery_TokenGenerationAndExpiration(t *testing.T) {
	cfg := &config.Config{
		RecoveryTokenTTL: 30 * time.Minute,
		RecoveryURLBase:  "http://localhost:8080/auth/password-recovery/reset?token=",
	}
	user := &models.Usuario{
		IDUsuario: "user-uuid-token-gen",
		Username:  "tokengen",
		Correo:    "tokengen@meloop.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"tokengen": user},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	recoveryRepo := &mockPasswordRecoveryRepository{tokens: make(map[string]*models.PasswordRecoveryToken)}
	emailService := &mockEmailService{}

	srv := createTestAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)

	before := time.Now()
	err := srv.RequestPasswordRecovery(context.Background(), &models.PasswordRecoveryRequest{
		Email: "tokengen@meloop.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(emailService.sentLinks) != 1 {
		t.Fatalf("expected 1 email link sent, got: %d", len(emailService.sentLinks))
	}
	rawToken := strings.TrimPrefix(emailService.sentLinks[0], cfg.RecoveryURLBase)
	if len(rawToken) != 64 { // 32 bytes hex = 64 hex characters
		t.Fatalf("expected 64-char hex raw token, got length %d (%s)", len(rawToken), rawToken)
	}

	for _, tokenRecord := range recoveryRepo.tokens {
		expectedExpiration := before.Add(30 * time.Minute)
		if tokenRecord.ExpiraEn.Before(expectedExpiration.Add(-2*time.Second)) || tokenRecord.ExpiraEn.After(expectedExpiration.Add(2*time.Second)) {
			t.Errorf("expected token expiration around %v, got %v", expectedExpiration, tokenRecord.ExpiraEn)
		}
	}
}

// 4. Token válido procesado correctamente
// 8. Cambio exitoso de contraseña
func TestResetPassword_ValidToken_Success(t *testing.T) {
	cfg := &config.Config{
		PasswordMinLength: 8,
		RecoveryTokenTTL:  30 * time.Minute,
		RecoveryURLBase:   "http://localhost:8080/auth/password-recovery/reset?token=",
	}
	initialHash, _ := bcrypt.GenerateFromPassword([]byte("OldP@ssword123"), bcrypt.DefaultCost)
	user := &models.Usuario{
		IDUsuario:      "user-uuid-reset",
		Username:       "alanuser",
		Correo:         "alan@meloop.com",
		ContrasenaHash: string(initialHash),
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"alanuser": user},
	}
	sessionRepo := &mockSessionRepository{
		sessions: map[string]*models.SessionUser{
			"active-session-1": {ID: "user-uuid-reset", Username: "alanuser", Email: "alan@meloop.com"},
		},
	}
	recoveryRepo := &mockPasswordRecoveryRepository{tokens: make(map[string]*models.PasswordRecoveryToken)}
	emailService := &mockEmailService{}

	srv := createTestAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)

	// 1. Solicitar recuperación
	err := srv.RequestPasswordRecovery(context.Background(), &models.PasswordRecoveryRequest{
		Email: "alan@meloop.com",
	})
	if err != nil {
		t.Fatalf("error in RequestPasswordRecovery: %v", err)
	}

	rawToken := strings.TrimPrefix(emailService.sentLinks[0], cfg.RecoveryURLBase)

	// 2. Restablecer con nueva contraseña válida
	newPassword := "NuevaP@ssw0rd2026"
	err = srv.ResetPassword(context.Background(), &models.ResetPasswordRequest{
		Token:       rawToken,
		NewPassword: newPassword,
	})
	if err != nil {
		t.Fatalf("expected successful ResetPassword, got: %v", err)
	}

	// 3. Verificar que la contraseña fue actualizada en hash bcrypt
	if user.ContrasenaHash == string(initialHash) {
		t.Error("expected contrasena_hash to be updated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.ContrasenaHash), []byte(newPassword)); err != nil {
		t.Errorf("bcrypt compare failed for new password: %v", err)
	}

	// 4. Verificar que las sesiones anteriores quedaron invalidadas
	if len(sessionRepo.sessions) != 0 {
		t.Errorf("expected active sessions to be invalidated, remaining: %d", len(sessionRepo.sessions))
	}
}

// 5. Token expirado
func TestResetPassword_ExpiredToken(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, RecoveryTokenTTL: 30 * time.Minute}
	user := &models.Usuario{
		IDUsuario: "user-uuid-expired",
		Username:  "expireduser",
		Correo:    "expired@meloop.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"expireduser": user},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}

	rawToken := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])

	recoveryRepo := &mockPasswordRecoveryRepository{
		tokens: map[string]*models.PasswordRecoveryToken{
			tokenHash: {
				IDRecuperacion: "rec-expired-1",
				IDUsuario:      "user-uuid-expired",
				TokenHash:      tokenHash,
				ExpiraEn:       time.Now().Add(-10 * time.Minute), // Expiró hace 10 minutos
				Usado:          false,
			},
		},
	}
	emailService := &mockEmailService{}

	srv := createTestAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)

	err := srv.ResetPassword(context.Background(), &models.ResetPasswordRequest{
		Token:       rawToken,
		NewPassword: "NewSecureP@ss123",
	})
	if !errors.Is(err, services.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got: %v", err)
	}
}

// 6. Token ya utilizado
// 10. Invalidación del token después de utilizarlo
// 12. Intento de reutilizar el enlace después de completar la recuperación
func TestResetPassword_AlreadyUsedToken_ReusingFails(t *testing.T) {
	cfg := &config.Config{
		PasswordMinLength: 8,
		RecoveryTokenTTL:  30 * time.Minute,
		RecoveryURLBase:   "http://localhost:8080/auth/password-recovery/reset?token=",
	}
	user := &models.Usuario{
		IDUsuario: "user-uuid-used",
		Username:  "useduser",
		Correo:    "used@meloop.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"useduser": user},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	recoveryRepo := &mockPasswordRecoveryRepository{tokens: make(map[string]*models.PasswordRecoveryToken)}
	emailService := &mockEmailService{}

	srv := createTestAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)

	err := srv.RequestPasswordRecovery(context.Background(), &models.PasswordRecoveryRequest{
		Email: "used@meloop.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawToken := strings.TrimPrefix(emailService.sentLinks[0], cfg.RecoveryURLBase)

	// Primer uso: Debe ser exitoso
	err = srv.ResetPassword(context.Background(), &models.ResetPasswordRequest{
		Token:       rawToken,
		NewPassword: "FirstResetP@ss123",
	})
	if err != nil {
		t.Fatalf("expected first reset to succeed, got: %v", err)
	}

	// Segundo uso: Debe ser rechazado como token ya utilizado (RN-04)
	err = srv.ResetPassword(context.Background(), &models.ResetPasswordRequest{
		Token:       rawToken,
		NewPassword: "SecondResetP@ss123",
	})
	if !errors.Is(err, services.ErrTokenAlreadyUsed) {
		t.Fatalf("expected ErrTokenAlreadyUsed when reusing token, got: %v", err)
	}
}

// 7. Token inválido o inexistente
func TestResetPassword_InvalidToken(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockUserRepository{users: make(map[string]*models.Usuario)}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	recoveryRepo := &mockPasswordRecoveryRepository{tokens: make(map[string]*models.PasswordRecoveryToken)}
	emailService := &mockEmailService{}

	srv := createTestAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)

	err := srv.ResetPassword(context.Background(), &models.ResetPasswordRequest{
		Token:       "token_que_no_existe_en_db",
		NewPassword: "NewValidP@ss123",
	})
	if !errors.Is(err, services.ErrTokenNotFound) {
		t.Fatalf("expected ErrTokenNotFound for invalid token, got: %v", err)
	}
}

// 9. Contraseña que incumple RN-19 (Política de contraseñas)
func TestResetPassword_RN19_PolicyViolations(t *testing.T) {
	cfg := &config.Config{
		PasswordMinLength: 8,
		RecoveryTokenTTL:  30 * time.Minute,
		RecoveryURLBase:   "http://localhost:8080/auth/password-recovery/reset?token=",
	}
	user := &models.Usuario{
		IDUsuario: "user-uuid-rn19",
		Username:  "alanpolo",
		Correo:    "alanpolo@meloop.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"alanpolo": user},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	recoveryRepo := &mockPasswordRecoveryRepository{tokens: make(map[string]*models.PasswordRecoveryToken)}
	emailService := &mockEmailService{}

	srv := createTestAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)

	_ = srv.RequestPasswordRecovery(context.Background(), &models.PasswordRecoveryRequest{
		Email: "alanpolo@meloop.com",
	})
	rawToken := strings.TrimPrefix(emailService.sentLinks[0], cfg.RecoveryURLBase)

	casosRN19 := []struct {
		nombre        string
		newPassword   string
		expectedIssue string
	}{
		{
			nombre:        "contraseña menor a 8 caracteres",
			newPassword:   "Ab1!x",
			expectedIssue: "too_short",
		},
		{
			nombre:        "contraseña mayor a 72 caracteres",
			newPassword:   strings.Repeat("A1b", 25), // 75 caracteres
			expectedIssue: "too_long",
		},
		{
			nombre:        "sin letras minúsculas",
			newPassword:   "PASSWORD123!",
			expectedIssue: "missing_lowercase",
		},
		{
			nombre:        "sin letras mayúsculas",
			newPassword:   "password123!",
			expectedIssue: "missing_uppercase",
		},
		{
			nombre:        "sin dígitos",
			newPassword:   "PasswordWithoutDigits!",
			expectedIssue: "missing_digit",
		},
	}

	for _, c := range casosRN19 {
		t.Run(c.nombre, func(t *testing.T) {
			err := srv.ResetPassword(context.Background(), &models.ResetPasswordRequest{
				Token:       rawToken,
				NewPassword: c.newPassword,
			})
			if err == nil {
				t.Fatalf("expected validation error for %s, got nil", c.nombre)
			}
			valErr, ok := err.(*services.ValidationError)
			if !ok {
				t.Fatalf("expected *ValidationError, got %T (%v)", err, err)
			}
			if valErr.Issue != c.expectedIssue {
				t.Errorf("expected issue '%s', got '%s' (message: %s)", c.expectedIssue, valErr.Issue, valErr.Message)
			}
		})
	}

	// Probar específicamente coincidencia con username
	userWithSpecialName := &models.Usuario{
		IDUsuario: "user-uuid-special-name",
		Username:  "AlanPolo123",
		Correo:    "special@meloop.com",
	}
	repo.users["AlanPolo123"] = userWithSpecialName
	_ = srv.RequestPasswordRecovery(context.Background(), &models.PasswordRecoveryRequest{
		Email: "special@meloop.com",
	})
	specialToken := strings.TrimPrefix(emailService.sentLinks[len(emailService.sentLinks)-1], cfg.RecoveryURLBase)

	err := srv.ResetPassword(context.Background(), &models.ResetPasswordRequest{
		Token:       specialToken,
		NewPassword: "AlanPolo123", // Cumple mayúsculas, minúsculas, dígitos, pero es igual al username
	})
	if err == nil {
		t.Fatal("expected error when new_password matches username")
	}
	valErr, ok := err.(*services.ValidationError)
	if !ok || valErr.Issue != "matches_username" {
		t.Errorf("expected issue 'matches_username', got: %v", err)
	}

	// Probar específicamente coincidencia con correo
	userWithEmailMatch := &models.Usuario{
		IDUsuario: "user-uuid-email-match",
		Username:  "SomeUser",
		Correo:    "AlanPolo123@meloop.com",
	}
	repo.users["SomeUser"] = userWithEmailMatch
	_ = srv.RequestPasswordRecovery(context.Background(), &models.PasswordRecoveryRequest{
		Email: "AlanPolo123@meloop.com",
	})
	emailMatchToken := strings.TrimPrefix(emailService.sentLinks[len(emailService.sentLinks)-1], cfg.RecoveryURLBase)

	err = srv.ResetPassword(context.Background(), &models.ResetPasswordRequest{
		Token:       emailMatchToken,
		NewPassword: "AlanPolo123@meloop.com", // Igual al correo
	})
	if err == nil {
		t.Fatal("expected error when new_password matches email")
	}
	valErr, ok = err.(*services.ValidationError)
	if !ok || valErr.Issue != "matches_email" {
		t.Errorf("expected issue 'matches_email', got: %v", err)
	}
}

// 11. Invalidación de sesiones anteriores
func TestResetPassword_InvalidatesActiveSessions(t *testing.T) {
	cfg := &config.Config{
		PasswordMinLength: 8,
		RecoveryTokenTTL:  30 * time.Minute,
		RecoveryURLBase:   "http://localhost:8080/auth/password-recovery/reset?token=",
	}
	user := &models.Usuario{
		IDUsuario: "user-uuid-multi-session",
		Username:  "multisession",
		Correo:    "multi@meloop.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"multisession": user},
	}
	sessionRepo := &mockSessionRepository{
		sessions: map[string]*models.SessionUser{
			"session-token-1":    {ID: "user-uuid-multi-session", Username: "multisession", Email: "multi@meloop.com"},
			"session-token-2":    {ID: "user-uuid-multi-session", Username: "multisession", Email: "multi@meloop.com"},
			"other-user-session": {ID: "user-uuid-other", Username: "other", Email: "other@meloop.com"},
		},
	}
	recoveryRepo := &mockPasswordRecoveryRepository{tokens: make(map[string]*models.PasswordRecoveryToken)}
	emailService := &mockEmailService{}

	srv := createTestAuthService(cfg, repo, sessionRepo, recoveryRepo, emailService)

	_ = srv.RequestPasswordRecovery(context.Background(), &models.PasswordRecoveryRequest{
		Email: "multi@meloop.com",
	})
	rawToken := strings.TrimPrefix(emailService.sentLinks[0], cfg.RecoveryURLBase)

	err := srv.ResetPassword(context.Background(), &models.ResetPasswordRequest{
		Token:       rawToken,
		NewPassword: "NewValidP@ssword2026",
	})
	if err != nil {
		t.Fatalf("expected reset to succeed, got: %v", err)
	}

	// Verificar que las sesiones del usuario afectado fueron eliminadas
	if _, exists := sessionRepo.sessions["session-token-1"]; exists {
		t.Error("expected session-token-1 to be deleted")
	}
	if _, exists := sessionRepo.sessions["session-token-2"]; exists {
		t.Error("expected session-token-2 to be deleted")
	}
	// La sesión de otro usuario debe preservarse
	if _, exists := sessionRepo.sessions["other-user-session"]; !exists {
		t.Error("expected other-user-session to remain active")
	}
}
