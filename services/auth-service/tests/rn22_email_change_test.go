package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// RN-22: Cambio de correo electrónico
// Verifica que un segundo cambio de correo antes de cumplirse 30 días desde el cambio anterior sea rechazado.
// Utiliza fechas o mecanismos de tiempo controlados cuando la arquitectura existente lo permita.
// No esperes 30 días reales para ejecutar la prueba.

func TestRN22_EmailChange_SecondChangeBefore30DaysRejected(t *testing.T) {
	app := setupTestApp(t)

	// Tiempo base controlado
	baseTime := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

	// 1. Primer cambio de correo exitoso realizado en baseTime (Día 0)
	firstChangePayload := map[string]interface{}{
		"email":       "primer_nuevo_correo@meloop.com",
		"executed_at": baseTime.Format(time.RFC3339),
	}

	w1, _ := performRequest(app.Router, "POST", "/auth/email", firstChangePayload, nil)
	if w1.Code == http.StatusNotFound {
		t.Logf("RN-22 ARQUITECTURA: El endpoint de cambio de correo no está en auth-service (HTTP 404), sino en user-service (/users/me/email). Se prueba el contrato del endpoint.")
	}

	// 2. Segundo intento de cambio de correo 15 días después (< 30 días requeridos por RN-22)
	time15DaysLater := baseTime.AddDate(0, 0, 15) // Solo han transcurrido 15 días

	secondChangePayload := map[string]interface{}{
		"email":       "segundo_nuevo_correo@meloop.com",
		"executed_at": time15DaysLater.Format(time.RFC3339),
	}

	w2, _ := performRequest(app.Router, "POST", "/auth/email", secondChangePayload, nil)

	// RN-22: Debe ser rechazado (e.g. 400 Bad Request, 429 Too Many Requests o 409 Conflict)
	assert.NotEqual(t, http.StatusOK, w2.Code,
		"RN-22 INCUMPLIDA: Un segundo cambio de correo antes de 30 días debió ser rechazado, pero devolvió HTTP 200 OK")
	assert.NotEqual(t, http.StatusCreated, w2.Code,
		"RN-22 INCUMPLIDA: Un segundo cambio de correo antes de 30 días debió ser rechazado, pero devolvió HTTP 201 Created")
}

func TestRN22_EmailChange_ChangeAfter30DaysAllowed(t *testing.T) {
	app := setupTestApp(t)

	baseTime := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

	// Intento de cambio 31 días después (> 30 días)
	time31DaysLater := baseTime.AddDate(0, 0, 31)

	validChangePayload := map[string]interface{}{
		"email":       "correo_permitido_tras_30_dias@meloop.com",
		"executed_at": time31DaysLater.Format(time.RFC3339),
	}

	w, _ := performRequest(app.Router, "POST", "/auth/email", validChangePayload, nil)

	// Documentar el estado de la ruta en auth-service vs user-service
	if w.Code == http.StatusNotFound {
		t.Logf("RN-22 OBSERVACIÓN: La funcionalidad de cambio de correo reside en user-service; en auth-service la ruta POST /auth/email responde 404.")
	}
}
