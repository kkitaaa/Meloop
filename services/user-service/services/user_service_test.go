package services_test

import (
	"context"
	"testing"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

// mockUserRepository es un mock en memoria de la interfaz UserRepository
type mockUserRepository struct {
	users              map[string]*models.Usuario
	getByIDFunc        func(ctx context.Context, id string) (*models.Usuario, error)
	getByUsernameFunc  func(ctx context.Context, username string) (*models.Usuario, error)
	getByEmailFunc     func(ctx context.Context, email string) (*models.Usuario, error)
	updateUsernameFunc func(ctx context.Context, id string, newUsername string) (*models.Usuario, error)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*models.Usuario, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	if u, ok := m.users[id]; ok {
		return u, nil
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

func (m *mockUserRepository) UpdateUsername(ctx context.Context, id string, newUsername string) (*models.Usuario, error) {
	if m.updateUsernameFunc != nil {
		return m.updateUsernameFunc(ctx, id, newUsername)
	}
	if u, ok := m.users[id]; ok {
		u.Username = newUsername
		return u, nil
	}
	return nil, nil
}

func TestGetAccount_Exitoso(t *testing.T) {
	user := &models.Usuario{
		IDUsuario:      "usr-123",
		Username:       "alanp",
		Correo:         "alan@example.com",
		ContrasenaHash: "hash_super_secreto",
		Experiencia:    100,
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-123": user},
	}
	srv := services.NewUserService(repo)

	res, err := srv.GetAccount(context.Background(), "usr-123")
	if err != nil {
		t.Fatalf("se esperaba éxito, pero ocurrió el error: %v", err)
	}

	if res.IDUsuario != "usr-123" {
		t.Errorf("se esperaba IDUsuario 'usr-123', se obtuvo '%s'", res.IDUsuario)
	}
	if res.Username != "alanp" {
		t.Errorf("se esperaba Username 'alanp', se obtuvo '%s'", res.Username)
	}
	if res.Correo != "alan@example.com" {
		t.Errorf("se esperaba Correo 'alan@example.com', se obtuvo '%s'", res.Correo)
	}
}

func TestGetAccount_UsuarioNoEncontrado(t *testing.T) {
	repo := &mockUserRepository{
		users: make(map[string]*models.Usuario),
	}
	srv := services.NewUserService(repo)

	_, err := srv.GetAccount(context.Background(), "no-existe")
	if err != services.ErrUserNotFound {
		t.Fatalf("se esperaba ErrUserNotFound, se obtuvo: %v", err)
	}
}

func TestUpdateUsername_Exitoso(t *testing.T) {
	user := &models.Usuario{
		IDUsuario: "usr-123",
		Username:  "antiguo",
		Correo:    "user@example.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-123": user},
	}
	srv := services.NewUserService(repo)

	req := &models.UpdateUsernameRequest{Username: "nuevo_username"}
	res, err := srv.UpdateUsername(context.Background(), "usr-123", req)
	if err != nil {
		t.Fatalf("se esperaba actualización exitosa, se obtuvo error: %v", err)
	}

	if res.Username != "nuevo_username" {
		t.Errorf("se esperaba nuevo username 'nuevo_username', se obtuvo '%s'", res.Username)
	}
	if user.Username != "nuevo_username" {
		t.Errorf("se esperaba que el repositorio persistiera el username, tiene: '%s'", user.Username)
	}
}

func TestUpdateUsername_MismoUsername(t *testing.T) {
	user := &models.Usuario{
		IDUsuario: "usr-123",
		Username:  "alanp",
		Correo:    "alan@example.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-123": user},
	}
	srv := services.NewUserService(repo)

	req := &models.UpdateUsernameRequest{Username: "alanp"}
	res, err := srv.UpdateUsername(context.Background(), "usr-123", req)
	if err != nil {
		t.Fatalf("se esperaba respuesta exitosa al no haber cambios, se obtuvo error: %v", err)
	}

	if res.Username != "alanp" {
		t.Errorf("se esperaba username 'alanp', se obtuvo '%s'", res.Username)
	}
}

func TestUpdateUsername_ErroresDeValidacion(t *testing.T) {
	user := &models.Usuario{
		IDUsuario: "usr-123",
		Username:  "alanp",
		Correo:    "alan@example.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-123": user},
	}
	srv := services.NewUserService(repo)

	casos := []struct {
		nombre        string
		req           *models.UpdateUsernameRequest
		expectedField string
		expectedIssue string
	}{
		{
			nombre:        "username vacío",
			req:           &models.UpdateUsernameRequest{Username: ""},
			expectedField: "username",
			expectedIssue: "required",
		},
		{
			nombre: "username supera 50 caracteres",
			req: &models.UpdateUsernameRequest{
				Username: "este_es_un_nombre_de_usuario_extremadamente_largo_que_supera_el_limite_de_cincuenta_caracteres",
			},
			expectedField: "username",
			expectedIssue: "too_long",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := srv.UpdateUsername(context.Background(), "usr-123", c.req)
			if err == nil {
				t.Fatal("se esperaba error de validación, pero fue nil")
			}
			valErr, ok := err.(*services.ValidationError)
			if !ok {
				t.Fatalf("se esperaba tipo *ValidationError, se obtuvo %T", err)
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

func TestUpdateUsername_UsernameDuplicado_RN06(t *testing.T) {
	user1 := &models.Usuario{IDUsuario: "usr-1", Username: "usuario1", Correo: "u1@test.com"}
	user2 := &models.Usuario{IDUsuario: "usr-2", Username: "usuario2", Correo: "u2@test.com"}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{
			"usr-1": user1,
			"usr-2": user2,
		},
	}
	srv := services.NewUserService(repo)

	// El usuario 1 intenta cambiar su username al username ya utilizado por el usuario 2
	req := &models.UpdateUsernameRequest{Username: "usuario2"}
	_, err := srv.UpdateUsername(context.Background(), "usr-1", req)
	if err != services.ErrUsernameExists {
		t.Fatalf("se esperaba ErrUsernameExists, se obtuvo: %v", err)
	}
}

func TestRequestEmailChange_Exitoso_ConservaCorreoActual_RN22(t *testing.T) {
	user := &models.Usuario{
		IDUsuario: "usr-123",
		Username:  "alanp",
		Correo:    "original@example.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-123": user},
	}
	srv := services.NewUserService(repo)

	req := &models.RequestEmailChangeRequest{Email: "nuevo@example.com"}
	res, err := srv.RequestEmailChange(context.Background(), "usr-123", req)
	if err != nil {
		t.Fatalf("se esperaba solicitud exitosa, se obtuvo: %v", err)
	}

	if res.PendingEmail != "nuevo@example.com" {
		t.Errorf("se esperaba PendingEmail 'nuevo@example.com', se obtuvo '%s'", res.PendingEmail)
	}

	// Regla RN-22: El correo actual en USUARIO.correo debe conservarse intacto
	if user.Correo != "original@example.com" {
		t.Errorf("el correo en la base de datos debió conservarse como 'original@example.com', pero cambió a '%s'", user.Correo)
	}
}

func TestRequestEmailChange_ErroresDeValidacion(t *testing.T) {
	user := &models.Usuario{
		IDUsuario: "usr-123",
		Username:  "alanp",
		Correo:    "original@example.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-123": user},
	}
	srv := services.NewUserService(repo)

	casos := []struct {
		nombre        string
		req           *models.RequestEmailChangeRequest
		expectedField string
		expectedIssue string
	}{
		{
			nombre:        "correo vacío",
			req:           &models.RequestEmailChangeRequest{Email: ""},
			expectedField: "email",
			expectedIssue: "required",
		},
		{
			nombre:        "formato de correo inválido",
			req:           &models.RequestEmailChangeRequest{Email: "correo_no_valido"},
			expectedField: "email",
			expectedIssue: "invalid_format",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := srv.RequestEmailChange(context.Background(), "usr-123", c.req)
			if err == nil {
				t.Fatal("se esperaba error de validación, pero fue nil")
			}
			valErr, ok := err.(*services.ValidationError)
			if !ok {
				t.Fatalf("se esperaba tipo *ValidationError, se obtuvo %T", err)
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

func TestRequestEmailChange_MismoCorreoActual(t *testing.T) {
	user := &models.Usuario{
		IDUsuario: "usr-123",
		Username:  "alanp",
		Correo:    "actual@example.com",
	}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{"usr-123": user},
	}
	srv := services.NewUserService(repo)

	req := &models.RequestEmailChangeRequest{Email: "actual@example.com"}
	_, err := srv.RequestEmailChange(context.Background(), "usr-123", req)
	if err != services.ErrEmailSameAsCurrent {
		t.Fatalf("se esperaba ErrEmailSameAsCurrent, se obtuvo: %v", err)
	}
}

func TestRequestEmailChange_CorreoDuplicado(t *testing.T) {
	user1 := &models.Usuario{IDUsuario: "usr-1", Username: "usuario1", Correo: "u1@test.com"}
	user2 := &models.Usuario{IDUsuario: "usr-2", Username: "usuario2", Correo: "u2@test.com"}
	repo := &mockUserRepository{
		users: map[string]*models.Usuario{
			"usr-1": user1,
			"usr-2": user2,
		},
	}
	srv := services.NewUserService(repo)

	// El usuario 1 intenta cambiar su correo al correo que ya tiene registrado el usuario 2
	req := &models.RequestEmailChangeRequest{Email: "u2@test.com"}
	_, err := srv.RequestEmailChange(context.Background(), "usr-1", req)
	if err != services.ErrEmailExists {
		t.Fatalf("se esperaba ErrEmailExists, se obtuvo: %v", err)
	}
}
