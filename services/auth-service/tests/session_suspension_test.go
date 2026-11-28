package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/repositories"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestValidateDeniesSuspensionAfterLogin(t *testing.T) {
	app := setupTestApp(t)
	password := "TestPassword123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	user := &models.Usuario{IDUsuario: "u1", Username: "ana", Correo: "ana@example.com", ContrasenaHash: string(hash)}
	app.UserRepo.Users[user.Username] = user
	login, err := app.AuthService.Login(context.Background(), &models.LoginRequest{Email: user.Correo, Password: password})
	require.NoError(t, err)
	headers := map[string]string{"Authorization": "Bearer " + login.SessionToken}

	active, _ := performRequest(app.Router, http.MethodGet, "/auth/validate", nil, headers)
	require.Equal(t, http.StatusOK, active.Code)
	app.UserRepo.GetRoleByIDFunc = func(_ context.Context, id string) (string, error) {
		require.Equal(t, user.IDUsuario, id)
		return "", repositories.ErrAccountSuspended
	}
	suspended, _ := performRequest(app.Router, http.MethodGet, "/auth/validate", nil, headers)
	require.Equal(t, http.StatusForbidden, suspended.Code)
	require.Contains(t, app.SessionRepo.Sessions, login.SessionToken)
}
