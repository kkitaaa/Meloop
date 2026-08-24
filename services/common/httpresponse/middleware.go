package httpresponse

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery(next http.Handler) http.Handler {
	return RecoveryWithLogger(slog.Default(), next)
}

func RecoveryWithLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic_recovered", "error", err, "method", r.Method, "path", r.URL.Path)

				InternalError(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// GinRecovery es un middleware para Gin que recupera pánicos y devuelve una respuesta estructurada
func GinRecovery() gin.HandlerFunc {
	return GinRecoveryWithLogger(slog.Default())
}

func GinRecoveryWithLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic_recovered", "error", err, "method", c.Request.Method, "path", c.Request.URL.Path)
				c.AbortWithStatusJSON(http.StatusInternalServerError, Response{
					Success: false,
					Data:    nil,
					Error: &ErrorBody{
						Code:    ErrInternal,
						Message: "Ha ocurrido un error interno",
					},
				})
			}
		}()
		c.Next()
	}
}
