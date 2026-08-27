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

func TestRegister_Success(t *testing.T) {
	cfg := &config.Config{PasswordMinLength: 8}
	repo := &mockAccountRepository{accounts: make(map[string]*models.Account)}
	srv := services.NewAuthService(cfg, repo)

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
	srv := services.NewAuthService(cfg, repo)

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
	srv := services.NewAuthService(cfg, repo)

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
	srv := services.NewAuthService(cfg, repo)

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
