package httpresponse

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ErrValidation   = "VALIDATION_ERROR"
	ErrUnauthorized = "UNAUTHORIZED"
	ErrForbidden    = "FORBIDDEN"
	ErrNotFound     = "NOT_FOUND"
	ErrConflict     = "CONFLICT"
	ErrInternal     = "INTERNAL_ERROR"
)

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, ErrValidation, message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, ErrUnauthorized, message)
}

func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, ErrForbidden, message)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, ErrNotFound, message)
}

func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, ErrConflict, message)
}

func InternalError(w http.ResponseWriter) {
	Error(
		w,
		http.StatusInternalServerError,
		ErrInternal,
		"Ha ocurrido un error interno",
	)
}

func BadRequestGin(c *gin.Context, message string) {
	ErrorGin(c, http.StatusBadRequest, ErrValidation, message)
}

func UnauthorizedGin(c *gin.Context, message string) {
	ErrorGin(c, http.StatusUnauthorized, ErrUnauthorized, message)
}

func ForbiddenGin(c *gin.Context, message string) {
	ErrorGin(c, http.StatusForbidden, ErrForbidden, message)
}

func NotFoundGin(c *gin.Context, message string) {
	ErrorGin(c, http.StatusNotFound, ErrNotFound, message)
}

func ConflictGin(c *gin.Context, message string) {
	ErrorGin(c, http.StatusConflict, ErrConflict, message)
}

func InternalErrorGin(c *gin.Context) {
	ErrorGin(
		c,
		http.StatusInternalServerError,
		ErrInternal,
		"Ha ocurrido un error interno",
	)
}
