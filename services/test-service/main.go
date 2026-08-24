package main

import (
	"log"
	"net/http"

	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	httpresponse.Success(w, http.StatusOK, map[string]string{
		"service": "test-service",
		"status":  "ok",
	})
}

func validationErrorHandler(w http.ResponseWriter, r *http.Request) {
	details := map[string]string{
		"field": "email",
		"issue": "invalid_format",
	}
	httpresponse.ErrorWithDetails(
		w,
		http.StatusBadRequest,
		httpresponse.ErrValidation,
		"El campo 'email' es obligatorio y debe ser un correo válido",
		details,
	)
}

func unauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	httpresponse.Unauthorized(w, "Token de autenticación vencido o inválido")
}

func forbiddenHandler(w http.ResponseWriter, r *http.Request) {
	httpresponse.Forbidden(w, "No tienes permisos para acceder a este recurso")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	httpresponse.NotFound(w, "El usuario con ID 123 no existe")
}

func internalErrorHandler(w http.ResponseWriter, r *http.Request) {
	httpresponse.InternalError(w)
}

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("¡Algo explotó en el servidor de pruebas!")
}

func main() {
	logger := logging.New("test-service")
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/validation-error", validationErrorHandler)
	mux.HandleFunc("/unauthorized", unauthorizedHandler)
	mux.HandleFunc("/forbidden", forbiddenHandler)
	mux.HandleFunc("/not-found", notFoundHandler)
	mux.HandleFunc("/internal-error", internalErrorHandler)
	mux.HandleFunc("/panic", panicHandler)

	// Envolver el enrutador con el middleware Recovery
	handler := httpresponse.RecoveryWithLogger(logger, logging.HTTPMiddleware(logger, mux))

	logger.Info("service_started", "port", 8081)

	if err := http.ListenAndServe(":8081", handler); err != nil {
		logger.Error("service_stopped", "error", err)
		log.Fatal(err)
	}
}
