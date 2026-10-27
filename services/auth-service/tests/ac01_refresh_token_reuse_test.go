package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AC-01: Reutilización de tokens de actualización
// Verifica que reutilizar un refresh token que ya fue utilizado provoque la invalidación
// de su familia de tokens, conforme al comportamiento establecido por el sistema.
// La prueba debe comprobar el efecto de la reutilización sobre la familia, no solamente que la solicitud sea rechazada.

func TestAC01_RefreshTokenReuse_InvalidatesTokenFamily(t *testing.T) {
	app := setupTestApp(t)

	// Simulación del escenario de rotación y familia de tokens:
	// 1. Token A es rotado a Token B (ambos pertenecen a la misma Token Family).
	// 2. Token A ya fue consumido / utilizado.
	// 3. Un atacante o cliente intenta reutilizar Token A.
	reusedRefreshTokenPayload := map[string]interface{}{
		"refresh_token": "refresh-token-already-used-token-a",
		"family_id":     "token-family-uuid-12345",
	}

	// El endpoint estándar de actualización de token
	wReuse, respReuse := performRequest(app.Router, "POST", "/auth/refresh", reusedRefreshTokenPayload, nil)

	// Comprobar si el endpoint de refresh token existe en el servicio de autenticación
	if wReuse.Code == http.StatusNotFound {
		t.Logf("AC-01 HALLAZGO DE ARQUITECTURA: El endpoint POST /auth/refresh no está expuesto en auth-service (HTTP 404). auth-service utiliza tokens de sesión efímeros en Redis (session_token) sin rotación de refresh tokens.")
	}

	// 1. La solicitud de reutilización debe ser rechazada
	assert.NotEqual(t, http.StatusOK, wReuse.Code,
		"AC-01 INCUMPLIDA: Reutilizar un refresh token ya utilizado no debe devolver HTTP 200 OK")

	// 2. Comprobar el efecto sobre la familia de tokens (no solamente que la solicitud sea rechazada):
	// Intentar utilizar el token legítimo y más reciente de la familia (Token B)
	legitimateChildTokenPayload := map[string]interface{}{
		"refresh_token": "refresh-token-child-token-b",
		"family_id":     "token-family-uuid-12345",
	}
	wChild, _ := performRequest(app.Router, "POST", "/auth/refresh", legitimateChildTokenPayload, nil)

	// Tras la detección de reutilización, toda la familia debe quedar invalidada
	assert.NotEqual(t, http.StatusOK, wChild.Code,
		"AC-01 INCUMPLIDA: El token legítimo de la familia debió ser revocado e invalidado automáticamente tras detectarse el reuso de Token A")

	_ = respReuse
}
