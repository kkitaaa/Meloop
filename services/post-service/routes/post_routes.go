package routes

import (
	"github.com/gin-gonic/gin"
	// Cambia esto al nombre real de tu módulo si es distinto en tu go.mod
	"github.com/meloop/post-service/controllers"
)

// SetupPostRoutes registra los endpoints de publicaciones en el router de Gin
func SetupPostRoutes(routerGroup *gin.RouterGroup, postCtrl *controllers.PostController) {

	// Agrupamos las rutas bajo "/posts"
	postRoutes := routerGroup.Group("/posts")
	{
		// RF-17: Ruta para que el autor modifique su contenido
		postRoutes.PUT("/:id", postCtrl.UpdatePost)

		// RF-18: Ruta para que el autor elimine su publicación
		postRoutes.DELETE("/:id", postCtrl.DeletePost)
	}
}
