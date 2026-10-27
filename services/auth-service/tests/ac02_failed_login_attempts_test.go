package tests

import (
	"net/http"
	"testing"

	"github.com/meloop/auth-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// AC-02: Intentos fallidos de inicio de sesión
// Verifica que el sexto intento fallido de inicio de sesión produzca HTTP 429 Too Many Requests,
// respetando la política de intentos y el mecanismo de bloqueo existentes.
// Aísla el estado entre pruebas para evitar que intentos anteriores alteren el resultado.

func TestAC02_FailedLoginAttempts_SixthAttemptReturns429(t *testing.T) {
	// Aislamiento completo de estado
	app := setupTestApp(t)

	// Crear usuario de prueba con credenciales válidas
	rawPassword := "CorrectPassword123!"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	user := &models.Usuario{
		IDUsuario:      "user-ac02-target",
		Username:       "lockout_user",
		Correo:         "lockout@meloop.com",
		ContrasenaHash: string(hashed),
	}
	app.UserRepo.Users[user.Username] = user

	invalidLoginPayload := models.LoginRequest{
		Email:    "lockout@meloop.com",
		Password: "IncorrectPassword!",
	}

	// 1. Ejecutar los primeros 5 intentos fallidos
	for i := 1; i <= 5; i++ {
		w, resp := performRequest(app.Router, "POST", "/auth/login", invalidLoginPayload, nil)
		require.Equal(t, http.StatusUnauthorized, w.Code,
			"Los intentos fallidos del 1 al 5 deben producir HTTP 401 Unauthorized (intento %d)", i)
		assert.False(t, resp["success"].(bool))
	}

	// 2. Ejecutar el sexto intento fallido
	w6, resp6 := performRequest(app.Router, "POST", "/auth/login", invalidLoginPayload, nil)

	// Criterio AC-02: El 6to intento fallido debe producir HTTP 429 Too Many Requests
	assert.Equal(t, http.StatusTooManyRequests, w6.Code,
		"AC-02 INCUMPLIDA: El sexto intento fallido debió producir HTTP 429 Too Many Requests, pero el sistema devolvió %d", w6.Code)

	if w6.Code == http.StatusTooManyRequests {
		require.NotNil(t, resp6)
		assert.False(t, resp6["success"].(bool))
		errObj, ok := resp6["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, []string{"RATE_LIMIT_EXCEEDED", "TOO_MANY_REQUESTS", "ACCOUNT_LOCKED"}, errObj["code"])
	}
}

func TestAC02_FailedLoginAttempts_StateIsolationBetweenTests(t *testing.T) {
	// Verificar que el estado no se filtre entre suites aisladas
	app := setupTestApp(t)

	rawPassword := "CorrectPassword123!"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	user := &models.Usuario{
		IDUsuario:      "user-ac02-isolated",
		Username:       "fresh_user",
		Correo:         "fresh@meloop.com",
		ContrasenaHash: string(hashed),
	}
	app.UserRepo.Users[user.Username] = user

	// En un estado nuevo y aislado, el primer intento no debe recibir 429
	invalidLoginPayload := models.LoginRequest{
		Email:    "fresh@meloop.com",
		Password: "WrongPassword!",
	}

	w, _ := performRequest(app.Router, "POST", "/auth/login", invalidLoginPayload, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"En un entorno aislado, el primer intento fallido debe responder con 401 y no estar bloqueado por pruebas previas")
}
