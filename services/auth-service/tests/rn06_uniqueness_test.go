package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/meloop/auth-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RN-06: Unicidad de correo y nombre de usuario
// Verifica por separado que:
// - Intentar registrar un correo electrónico ya existente produzca HTTP 409 Conflict.
// - Intentar registrar un nombre de usuario ya existente produzca HTTP 409 Conflict.
// - Las respuestas no revelen datos personales ni información de la cuenta existente.
// Comprueba el contenido real de las respuestas; no basta con verificar únicamente el código HTTP.

func TestRN06_DuplicateEmail_Produces409Conflict_AndNoDataLeak(t *testing.T) {
	app := setupTestApp(t)

	// Pre-registrar un usuario existente con datos sensibles
	existingUser := &models.Usuario{
		IDUsuario:      "sensitive-uuid-12345-email-target",
		Username:       "usuario_existente",
		Correo:         "correo_existente@meloop.com",
		ContrasenaHash: "$2a$10$SensitivePasswordHashMustNeverBeLeakedInResponses",
		Experiencia:    500,
	}
	app.UserRepo.Users[existingUser.Username] = existingUser

	// Intentar registrar un nuevo usuario con nombre de usuario distinto pero correo duplicado
	newRegistrationPayload := models.RegisterRequest{
		Username: "nuevo_usuario_diferente",
		Email:    "correo_existente@meloop.com", // Mismo correo que el ya existente
		Password: "SecurePassword123!",
	}

	w, resp := performRequest(app.Router, "POST", "/auth/register", newRegistrationPayload, nil)

	// 1. Verificar HTTP 409 Conflict
	require.Equal(t, http.StatusConflict, w.Code, "debe responder estrictamente con HTTP 409 Conflict ante correo duplicado")

	// 2. Verificar estructura y código de error
	require.NotNil(t, resp, "el cuerpo de respuesta JSON no debe estar vacío")
	assert.False(t, resp["success"].(bool), "el campo 'success' debe ser false")
	assert.Nil(t, resp["data"], "el campo 'data' debe ser nulo")

	errObj, ok := resp["error"].(map[string]interface{})
	require.True(t, ok, "el cuerpo debe contener el objeto 'error'")
	assert.Equal(t, "CONFLICT", errObj["code"], "el código de error debe ser CONFLICT")

	// 3. Comprobación exhaustiva del contenido real: NO revelar datos personales ni información de la cuenta existente
	rawResponseBody := w.Body.String()

	// No debe exponer el nombre de usuario de la cuenta existente
	assert.False(t, strings.Contains(rawResponseBody, "usuario_existente"),
		"VULNERABILIDAD: La respuesta no debe revelar el nombre de usuario de la cuenta existente")

	// No debe exponer el identificador interno / UUID de la cuenta existente
	assert.False(t, strings.Contains(rawResponseBody, "sensitive-uuid-12345"),
		"VULNERABILIDAD: La respuesta no debe revelar el ID de usuario de la cuenta existente")

	// No debe exponer hashes de contraseña o información de contraseñas
	assert.False(t, strings.Contains(rawResponseBody, "$2a$"),
		"VULNERABILIDAD: La respuesta no debe contener fragmentos de hashes bcrypt")
	assert.False(t, strings.Contains(rawResponseBody, "SensitivePasswordHash"),
		"VULNERABILIDAD: La respuesta no debe revelar hashes o secretos de contraseñas")

	// No debe incluir campos sensibles en el objeto JSON
	assert.NotContains(t, errObj, "user_id")
	assert.NotContains(t, errObj, "username")
	assert.NotContains(t, errObj, "id_usuario")
	assert.NotContains(t, errObj, "contrasena_hash")
	assert.NotContains(t, errObj, "password_hash")
}

func TestRN06_DuplicateUsername_Produces409Conflict_AndNoDataLeak(t *testing.T) {
	app := setupTestApp(t)

	// Pre-registrar un usuario existente con datos sensibles
	existingUser := &models.Usuario{
		IDUsuario:      "sensitive-uuid-67890-username-target",
		Username:       "usuario_duplicado_test",
		Correo:         "correo_privado_existente@meloop.com",
		ContrasenaHash: "$2a$10$VerySecretHashExistingUserPassword2026",
		Experiencia:    1200,
	}
	app.UserRepo.Users[existingUser.Username] = existingUser

	// Intentar registrar un nuevo usuario con el mismo username pero correo diferente
	newRegistrationPayload := models.RegisterRequest{
		Username: "usuario_duplicado_test", // Mismo username que el ya existente
		Email:    "correo_completamente_nuevo@meloop.com",
		Password: "SecurePassword123!",
	}

	w, resp := performRequest(app.Router, "POST", "/auth/register", newRegistrationPayload, nil)

	// 1. Verificar HTTP 409 Conflict
	require.Equal(t, http.StatusConflict, w.Code, "debe responder estrictamente con HTTP 409 Conflict ante username duplicado")

	// 2. Verificar estructura y código de error
	require.NotNil(t, resp, "el cuerpo de respuesta JSON no debe estar vacío")
	assert.False(t, resp["success"].(bool), "el campo 'success' debe ser false")
	assert.Nil(t, resp["data"], "el campo 'data' debe ser nulo")

	errObj, ok := resp["error"].(map[string]interface{})
	require.True(t, ok, "el cuerpo debe contener el objeto 'error'")
	assert.Equal(t, "CONFLICT", errObj["code"], "el código de error debe ser CONFLICT")

	// 3. Comprobación exhaustiva del contenido real: NO revelar datos personales ni información de la cuenta existente
	rawResponseBody := w.Body.String()

	// No debe exponer el correo electrónico privado asociado a la cuenta existente
	assert.False(t, strings.Contains(rawResponseBody, "correo_privado_existente@meloop.com"),
		"VULNERABILIDAD: La respuesta revela el correo electrónico privado de la cuenta existente")

	// No debe exponer el ID interno del usuario
	assert.False(t, strings.Contains(rawResponseBody, "sensitive-uuid-67890"),
		"VULNERABILIDAD: La respuesta revela el ID interno de la cuenta existente")

	// No debe filtrar hashes de contraseña
	assert.False(t, strings.Contains(rawResponseBody, "$2a$"),
		"VULNERABILIDAD: La respuesta no debe contener fragmentos de hashes")
	assert.False(t, strings.Contains(rawResponseBody, "VerySecretHash"),
		"VULNERABILIDAD: La respuesta no debe exponer hashes de contraseñas")

	// Verificar ausencia de claves de datos de usuario en la respuesta de error
	var rawMap map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &rawMap)
	require.NoError(t, err)

	assert.Nil(t, rawMap["data"])
	assert.NotContains(t, errObj, "email")
	assert.NotContains(t, errObj, "correo")
	assert.NotContains(t, errObj, "user_id")
	assert.NotContains(t, errObj, "id_usuario")
}
