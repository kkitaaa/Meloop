package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/meloop/auth-service/config"
	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/services"
)

type mockAccountRepository struct {
	accounts          map[string]*models.Account
	getByUsernameFunc func(ctx context.Context, username string) (*models.Account, error)
	getByEmailFunc    func(ctx context.Context, email string) (*models.Account, error)
	createFunc        func(ctx context.Context, account *models.Account) error
}

func (m *mockAccountRepository) GetByUsername(ctx context.Context, username string) (*models.Account, error) {
	if m.getByUsernameFunc != nil {
		return m.getByUsernameFunc(ctx, username)
	}
	for _, acc := range m.accounts {
		if acc.Username == username {
			return acc, nil
		}
	}
	return nil, nil
}

func (m *mockAccountRepository) GetByEmail(ctx context.Context, email string) (*models.Account, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	for _, acc := range m.accounts {
		if acc.Email == email {
			return acc, nil
		}
	}
	return nil, nil
}

func (m *mockAccountRepository) Create(ctx context.Context, account *models.Account) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, account)
	}
	account.ID = "mock-uuid-123"
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()
	m.accounts[account.Username] = account
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
	repo := &mockAccountRepository{accounts: make(map[string]*models.Account)}
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

	if res.ID == "" {
		t.Error("expected valid ID in response")
	}
	if res.Username != req.Username {
		t.Errorf("expected username %s, got %s", req.Username, res.Username)
	}
	if res.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, res.Email)
	}

	// Verify that password hash was saved in repository and password is correct
	savedAcc := repo.accounts[req.Username]
	if savedAcc == nil {
		t.Fatal("expected account to be persisted in repo")
	}
	if savedAcc.PasswordHash == "" {
		t.Error("expected password hash to be set")
	}
	if savedAcc.PasswordHash == req.Password {
		t.Error("expected password NOT to be stored in plain text")
	}

	err = bcrypt.CompareHashAndPassword([]byte(savedAcc.PasswordHash), []byte(req.Password))
	if err != nil {
		t.Errorf("password verification failed: %v", err)
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockAccountRepository{accounts: make(map[string]*models.Account)}
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
				Email:    "invalid-email",
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
				Password: "password123password123password123password123password123password123password123", // > 72 chars
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

func TestRegister_DuplicateChecks(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockAccountRepository{
		accounts: map[string]*models.Account{
			"existinguser": {
				ID:           "uuid-existing-1",
				Username:     "existinguser",
				Email:        "existing@example.com",
				PasswordHash: "somehash",
			},
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := services.NewAuthService(cfg, repo, sessionRepo)

	// Test duplicate username
	req1 := &models.RegisterRequest{
		Username: "existinguser",
		Email:    "newemail@example.com",
		Password: "password123",
	}
	_, err := srv.Register(context.Background(), req1)
	if !errors.Is(err, services.ErrUsernameExists) {
		t.Errorf("expected ErrUsernameExists, got %v", err)
	}

	// Test duplicate email
	req2 := &models.RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
	}
	_, err = srv.Register(context.Background(), req2)
	if !errors.Is(err, services.ErrEmailExists) {
		t.Errorf("expected ErrEmailExists, got %v", err)
	}
}

func TestRegister_DatabaseUniqueViolationRace(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockAccountRepository{
		accounts: make(map[string]*models.Account),
		createFunc: func(ctx context.Context, account *models.Account) error {
			// Simulate unique violation error (code 23505) from database driver
			return errors.New("ERROR: duplicate key value violates unique constraint \"idx_accounts_username\" (SQLSTATE 23505)")
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := services.NewAuthService(cfg, repo, sessionRepo)

	req := &models.RegisterRequest{
		Username: "duplicateuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	_, err := srv.Register(context.Background(), req)
	if !errors.Is(err, services.ErrUsernameExists) {
		t.Errorf("expected ErrUsernameExists from database race mapping, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("meloop123"), bcrypt.DefaultCost)
	repo := &mockAccountRepository{
		accounts: map[string]*models.Account{
			"alan": {
				ID:           "user-uuid-1",
				Username:     "alan",
				Email:        "alan@meloop.com",
				PasswordHash: string(hashedPassword),
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

func TestLogin_InvalidCredentials(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("meloop123"), bcrypt.DefaultCost)
	repo := &mockAccountRepository{
		accounts: map[string]*models.Account{
			"alan": {
				ID:           "user-uuid-1",
				Username:     "alan",
				Email:        "alan@meloop.com",
				PasswordHash: string(hashedPassword),
			},
		},
	}
	sessionRepo := &mockSessionRepository{sessions: make(map[string]*models.SessionUser)}
	srv := services.NewAuthService(cfg, repo, sessionRepo)

	// Case 1: Email not found
	req1 := &models.LoginRequest{
		Email:    "wrong@meloop.com",
		Password: "meloop123",
	}
	_, err := srv.Login(context.Background(), req1)
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for non-existent email, got %v", err)
	}

	// Case 2: Wrong password
	req2 := &models.LoginRequest{
		Email:    "alan@meloop.com",
		Password: "wrongpassword",
	}
	_, err = srv.Login(context.Background(), req2)
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for incorrect password, got %v", err)
	}
}

func TestLogout_Success(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	repo := &mockAccountRepository{accounts: make(map[string]*models.Account)}
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

func TestValidateSession(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8, SessionTTL: 24 * time.Hour}
	repo := &mockAccountRepository{accounts: make(map[string]*models.Account)}
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

	// Case 2: Invalid/non-existent token
	_, err = srv.ValidateSession(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for invalid session token, got nil")
	}
}
