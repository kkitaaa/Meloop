package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RN-08: Edad mínima
// Verifica que:
// - Una persona de 12 años y 364 días no pueda registrarse.
// - Una persona que acaba de cumplir 13 años sí pueda registrarse.
// Utiliza fechas de nacimiento controladas para que los resultados sean reproducibles.
// Respeta el método de cálculo de edad y los contratos existentes del sistema.

func TestRN08_MinimumAge_UnderageUser_12Years364Days_CannotRegister(t *testing.T) {
	app := setupTestApp(t)

	// Fecha base controlada relativa a la fecha actual para reproducibilidad universal
	now := time.Now().UTC()

	// Cálculo exacto: 13 años menos 1 día = 12 años y 364 días (cumple 13 años mañana)
	birthdateUnderage := now.AddDate(-13, 0, 1)
	birthdateStr := birthdateUnderage.Format("2006-01-02")

	payload := map[string]interface{}{
		"username":         "underage_teen",
		"email":            "underage@meloop.com",
		"password":         "ValidPass123!",
		"fecha_nacimiento": birthdateStr,
		"birth_date":       birthdateStr,
	}

	w, resp := performRequest(app.Router, "POST", "/auth/register", payload, nil)

	// Según RN-08, el sistema debe rechazar el registro de usuarios que no hayan cumplido los 13 años con 400 Bad Request
	require.Equal(t, http.StatusBadRequest, w.Code,
		"RN-08: Una persona de 12 años y 364 días debe ser rechazada con HTTP 400 Bad Request")

	require.NotNil(t, resp)
	assert.False(t, resp["success"].(bool))
	errObj, ok := resp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", errObj["code"])

	details, ok := errObj["details"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "fecha_nacimiento", details["field"])
	assert.Equal(t, "underage", details["issue"])
}

func TestRN08_MinimumAge_ValidUser_Exactly13Years_CanRegister(t *testing.T) {
	app := setupTestApp(t)

	now := time.Now().UTC()

	// Cálculo exacto: exactamente 13 años cumplidos hoy
	birthdateValid := now.AddDate(-13, 0, 0)
	birthdateStr := birthdateValid.Format("2006-01-02")

	payload := map[string]interface{}{
		"username":         "newly_turned_13",
		"email":            "turned13@meloop.com",
		"password":         "ValidPass123!",
		"fecha_nacimiento": birthdateStr,
		"birth_date":       birthdateStr,
	}

	w, resp := performRequest(app.Router, "POST", "/auth/register", payload, nil)

	// Según RN-08, una persona que acaba de cumplir 13 años sí debe poder registrarse
	require.Equal(t, http.StatusCreated, w.Code,
		"RN-08: Una persona que acaba de cumplir 13 años debe ser registrada exitosamente (HTTP 201)")

	require.NotNil(t, resp)
	assert.True(t, resp["success"].(bool))
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "newly_turned_13", data["username"])
	assert.Equal(t, "turned13@meloop.com", data["email"])
}
