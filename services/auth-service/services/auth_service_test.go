package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/meloop/auth-service/config"
	"github.com/meloop/auth-service/models"
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
	sessions   map[string]*models.SessionUser
	createFunc func(ctx context.Context, token string, user *models.SessionUser, ttl time.Duration) error
	getFunc    func(ctx context.Context, token string) (*models.SessionUser, error)
	deleteFunc func(ctx context.Context, token string) error
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

func TestRegister_Success(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockUserRepository{users: make(map[string]*models.Usuario)}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srvUsername := services.NewAuthService(cfg, repoUsernameRace, sessionRepo)
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
	srvEmail := services.NewAuthService(cfg, repoEmailRace, sessionRepo)
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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

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
	srv := services.NewAuthService(cfg, repo, sessionRepo)

	req := &models.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newPassword456",
	}

	err := srv.ChangePassword(context.Background(), "user-inexistente", req)
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials cuando el usuario no existe, se obtuvo: %v", err)
	}
}
