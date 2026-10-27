package tests

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/meloop/auth-service/models"
	"github.com/meloop/auth-service/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RN-19: Validación de contraseñas
// Verifica que el sistema rechace:
// - Contraseñas de 7 caracteres.
// - Contraseñas de 73 caracteres.
// - Contraseñas sin mayúsculas.
// - Contraseñas sin minúsculas.
// - Contraseñas sin dígitos.
// - Contraseñas iguales al nombre de usuario.
// - Contraseñas iguales al correo electrónico.
// Verifica que acepte contraseñas válidas de exactamente 8 y exactamente 72 caracteres.

func TestRN19_ValidatePasswordPolicy_RejectedCases(t *testing.T) {
	username := "AlanPolo123"
	email := "AlanPolo123@meloop.com"
	minLength := 8

	testCases := []struct {
		name          string
		password      string
		expectedIssue string
		description   string
	}{
		{
			name:          "contraseña de 7 caracteres",
			password:      "Ab1!xyz", // longitud 7 (< 8)
			expectedIssue: "too_short",
			description:   "Debe rechazar contraseñas con longitud menor al mínimo requerido (7 caracteres)",
		},
		{
			name:          "contraseña de 73 caracteres",
			password:      "Ab1" + strings.Repeat("x", 70), // longitud 73 (> 72)
			expectedIssue: "too_long",
			description:   "Debe rechazar contraseñas que superen el límite máximo de 72 caracteres",
		},
		{
			name:          "contraseña sin mayúsculas",
			password:      "password123!", // sin mayúsculas
			expectedIssue: "missing_uppercase",
			description:   "Debe rechazar contraseñas sin al menos una letra mayúscula",
		},
		{
			name:          "contraseña sin minúsculas",
			password:      "PASSWORD123!", // sin minúsculas
			expectedIssue: "missing_lowercase",
			description:   "Debe rechazar contraseñas sin al menos una letra minúscula",
		},
		{
			name:          "contraseña sin dígitos",
			password:      "PasswordWithoutDigits!", // sin dígitos
			expectedIssue: "missing_digit",
			description:   "Debe rechazar contraseñas sin al menos un dígito",
		},
		{
			name:          "contraseña igual al nombre de usuario",
			password:      "AlanPolo123", // igual a username (cumple mayús, minús, dígitos)
			expectedIssue: "matches_username",
			description:   "Debe rechazar contraseñas idénticas al nombre de usuario",
		},
		{
			name:          "contraseña igual al correo electrónico",
			password:      "AlanPolo123@meloop.com", // igual a email (cumple mayús, minús, dígitos)
			expectedIssue: "matches_email",
			description:   "Debe rechazar contraseñas idénticas al correo electrónico registrado",
		},
	}


	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := services.ValidatePasswordPolicy(tc.password, "password", minLength, username, email)
			require.Error(t, err, tc.description)

			var valErr *services.ValidationError
			require.ErrorAs(t, err, &valErr, "el error devuelto debe ser de tipo *services.ValidationError")
			assert.Equal(t, tc.expectedIssue, valErr.Issue, "el código de issue debe coincidir exactamente")
		})
	}
}

func TestRN19_ValidatePasswordPolicy_AcceptedBoundaryCases(t *testing.T) {
	username := "validusername"
	email := "validuser@meloop.com"
	minLength := 8

	boundaryCases := []struct {
		name        string
		password    string
		description string
	}{
		{
			name:        "contraseña válida de exactamente 8 caracteres",
			password:    "Abcdef1!", // exactamente 8 caracteres con mayúscula, minúscula, dígito
			description: "Debe aceptar contraseñas válidas que alcancen exactamente el mínimo de 8 caracteres",
		},
		{
			name:        "contraseña válida de exactamente 72 caracteres",
			password:    "Ab1" + strings.Repeat("x", 69), // exactamente 72 caracteres (3 + 69 = 72)
			description: "Debe aceptar contraseñas válidas que alcancen exactamente el máximo permitido de 72 caracteres",
		},
	}

	for _, bc := range boundaryCases {
		t.Run(bc.name, func(t *testing.T) {
			assert.Equal(t, 8, len(boundaryCases[0].password), "la longitud de prueba debe ser exactamente 8")
			assert.Equal(t, 72, len(boundaryCases[1].password), "la longitud de prueba debe ser exactamente 72")

			err := services.ValidatePasswordPolicy(bc.password, "password", minLength, username, email)
			assert.NoError(t, err, bc.description)
		})
	}
}

func TestRN19_HTTP_ResetPassword_AllPolicyCases(t *testing.T) {
	app := setupTestApp(t)

	// Pre-configurar un usuario y token de recuperación válido para probar el contrato HTTP
	testUser := &models.Usuario{
		IDUsuario: "user-rn19-http",
		Username:  "AlanPolo123",
		Correo:    "AlanPolo123@meloop.com",
	}
	app.UserRepo.Users[testUser.Username] = testUser

	rawToken := "valid-recovery-token-rn19-test-123456"

	// Mock del repositorio de recuperación para aceptar este token
	app.RecoveryRepo.GetTokenByHashFunc = func(ctx context.Context, hash string) (*models.PasswordRecoveryToken, error) {
		return &models.PasswordRecoveryToken{
			IDRecuperacion: "rec-rn19-1",
			IDUsuario:      testUser.IDUsuario,
			TokenHash:      hash,
			ExpiraEn:       time.Now().Add(30 * time.Minute),
			Usado:          false,
		}, nil
	}

	cases := []struct {
		name          string
		password      string
		expectedCode  int
		expectedIssue string
		isSuccess     bool
	}{
		{
			name:          "HTTP rechaza 7 caracteres",
			password:      "Ab1!xyz",
			expectedCode:  http.StatusBadRequest,
			expectedIssue: "too_short",
			isSuccess:     false,
		},
		{
			name:          "HTTP rechaza 73 caracteres",
			password:      "Ab1" + strings.Repeat("x", 70),
			expectedCode:  http.StatusBadRequest,
			expectedIssue: "too_long",
			isSuccess:     false,
		},
		{
			name:          "HTTP rechaza sin mayúsculas",
			password:      "password123!",
			expectedCode:  http.StatusBadRequest,
			expectedIssue: "missing_uppercase",
			isSuccess:     false,
		},
		{
			name:          "HTTP rechaza sin minúsculas",
			password:      "PASSWORD123!",
			expectedCode:  http.StatusBadRequest,
			expectedIssue: "missing_lowercase",
			isSuccess:     false,
		},
		{
			name:          "HTTP rechaza sin dígitos",
			password:      "PasswordWithoutDigits!",
			expectedCode:  http.StatusBadRequest,
			expectedIssue: "missing_digit",
			isSuccess:     false,
		},
		{
			name:          "HTTP rechaza igual al nombre de usuario",
			password:      "AlanPolo123",
			expectedCode:  http.StatusBadRequest,
			expectedIssue: "matches_username",
			isSuccess:     false,
		},
		{
			name:          "HTTP rechaza igual al correo electrónico",
			password:      "AlanPolo123@meloop.com",
			expectedCode:  http.StatusBadRequest,
			expectedIssue: "matches_email",
			isSuccess:     false,
		},
		{
			name:         "HTTP acepta exactamente 8 caracteres válidos",
			password:     "Abcdef1!",
			expectedCode: http.StatusOK,
			isSuccess:    true,
		},
		{
			name:         "HTTP acepta exactamente 72 caracteres válidos",
			password:     "Ab1" + strings.Repeat("x", 69),
			expectedCode: http.StatusOK,
			isSuccess:    true,
		},
	}


	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			payload := map[string]string{
				"token":        rawToken,
				"new_password": c.password,
			}

			w, resp := performRequest(app.Router, "POST", "/auth/password-recovery/reset", payload, nil)

			require.Equal(t, c.expectedCode, w.Code, "el código de respuesta HTTP debe coincidir con lo esperado")

			if c.isSuccess {
				assert.True(t, resp["success"].(bool), "la respuesta debe indicar éxito")
			} else {
				assert.False(t, resp["success"].(bool), "la respuesta debe indicar fallo")
				errObj, ok := resp["error"].(map[string]interface{})
				require.True(t, ok, "debe contener el objeto 'error'")
				assert.Equal(t, "VALIDATION_ERROR", errObj["code"])

				details, ok := errObj["details"].(map[string]interface{})
				require.True(t, ok, "debe contener detalles de validación")
				assert.Equal(t, c.expectedIssue, details["issue"])
			}
		})
	}
}
