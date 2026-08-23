package httpresponse

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)

				InternalError(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// GinRecovery es un middleware para Gin que recupera pánicos y devuelve una respuesta estructurada
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered (Gin): %v", err)
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
